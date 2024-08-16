package temporalworkflows

import "github.com/google/uuid"

func getWorkflowReferenceID(workflowReference uuid.UUID) uuid.UUID {
	return uuid.NewSHA1(
		uuid.NameSpaceDNS,
		[]byte(workflowReference.String()),
	)
}
