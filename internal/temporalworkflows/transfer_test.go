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

func TestTransfer(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(transfersTestSuite))
}

func (s *transfersTestSuite) TestTransferWorkflow() {
	s.Run("Transfer", func() {
		var transactionOperations *activities.TransactionOperations
		var redisActivity *activities.Mutex

		activityResponse := &activities.TransferResult{}

		s.env.OnActivity(redisActivity.AcquireLock, mock.Anything, mock.Anything).Return(nil)

		s.env.OnActivity(
			transactionOperations.Transfer,
			mock.Anything,
			mock.Anything,
		).Return(activityResponse, nil)

		s.env.ExecuteWorkflow(Transfer, &TransferParams{})

		s.True(s.env.IsWorkflowCompleted())
		s.NoError(s.env.GetWorkflowError())
	})
}
