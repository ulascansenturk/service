package temporalworkflows

import (
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
	"ulascansenturk/service/internal/temporalworkflows/activities"

	temporalMocks "go.temporal.io/sdk/mocks"
	"go.temporal.io/sdk/testsuite"
)

type transfersTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite

	temporalService temporalMocks.Client
	env             *testsuite.TestWorkflowEnvironment
}

func (s *transfersTestSuite) SetupSubTest() {
	s.env = s.NewTestWorkflowEnvironment()
	s.temporalService = temporalMocks.Client{}

	s.env.RegisterWorkflow(Transfer)
}

func (s *transfersTestSuite) TearDownSubTest() {
	s.env.AssertExpectations(s.T())
}

func TestInvalidateOutdatedVerifications(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(transfersTestSuite))
}

func (s *transfersTestSuite) TestTransfer() {
	s.Run("Transfer", func() {
		var transactionOperations activities.TransactionOperations

		activityResponse := &activities.TransferResult{}

		s.env.OnActivity(
			transactionOperations.Transfer,
			mock.Anything,
			&TransferParams{},
		).Return(activityResponse, nil)

		s.env.ExecuteWorkflow(Transfer)

		s.True(s.env.IsWorkflowCompleted())
		s.NoError(s.env.GetWorkflowError())
	})
}
