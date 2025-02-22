// Copyright © 2019 Banzai Cloud
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package workflow

import (
	"context"

	"emperror.dev/errors"
	"github.com/aws/aws-sdk-go/service/ec2"
	"go.uber.org/cadence/activity"
	zapadapter "logur.dev/adapter/zap"

	"github.com/banzaicloud/pipeline/internal/cluster/infrastructure/aws/awsworkflow"
	pkgEC2 "github.com/banzaicloud/pipeline/pkg/providers/amazon/ec2"
)

const DeleteOrphanNICActivityName = "eks-delete-orphan-nic"

// DeleteOrphanNICActivity responsible for deleting asg
type DeleteOrphanNICActivity struct {
	awsSessionFactory *awsworkflow.AWSSessionFactory
}

type DeleteOrphanNICActivityInput struct {
	EKSActivityInput
	NicID string
}

type DeleteOrphanNICActivityOutput struct{}

//   DeleteOrphanNICActivity instantiates a new DeleteOrphanNICActivity
func NewDeleteOrphanNICActivity(awsSessionFactory *awsworkflow.AWSSessionFactory) *DeleteOrphanNICActivity {
	return &DeleteOrphanNICActivity{
		awsSessionFactory: awsSessionFactory,
	}
}

func (a *DeleteOrphanNICActivity) Execute(ctx context.Context, input DeleteOrphanNICActivityInput) error {
	logger := activity.GetLogger(ctx).Sugar().With(
		"organization", input.OrganizationID,
		"cluster", input.ClusterName,
	)

	awsSession, err := a.awsSessionFactory.New(input.OrganizationID, input.SecretID, input.Region)
	if err = errors.WrapIf(err, "failed to create AWS session"); err != nil {
		return err
	}

	netSvc := pkgEC2.NewNetworkSvc(ec2.New(awsSession), zapadapter.New(logger.Desugar()))
	logger.Info("deleting network interface", map[string]interface{}{"nic": input.NicID})

	return netSvc.DeleteNetworkInterface(input.NicID)
}
