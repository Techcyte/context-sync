package util

import (
	"tcs/internal/model"

	"github.com/google/uuid"
)

func NewSubRequestMessage(application string, replaceExistingClient bool) model.Message {
	info := model.ConnectionInfo{
		Version:              1,
		Application:          application,
		ReplaceExitingClient: &replaceExistingClient,
	}

	return model.Message{
		Kind: model.SubscriptionRequest,
		Info: &info,
	}
}

func NewSubAcceptMessage(application string, timeout *float64, caseNumber string) model.Message {
	info := model.ConnectionInfo{
		Version:     1,
		Application: application,
		Timeout:     timeout,
	}

	var context []model.ContextItem
	if caseNumber != "" {
		context = []model.ContextItem{
			{Key: model.CaseNumber, Value: caseNumber},
		}
	}

	return model.Message{
		Kind:    model.SubscriptionAccept,
		Info:    &info,
		Context: context,
	}
}

func NewSubRejectMessage(application string, timeout *float64, reason string, status model.StatusCode) model.Message {
	info := model.ConnectionInfo{
		Version:     1,
		Application: application,
		Timeout:     timeout,
	}

	rejection := model.MessageRejection{
		Reason: reason,
		Status: status,
	}

	return model.Message{
		Kind:      model.SubscriptionReject,
		Info:      &info,
		Rejection: &rejection,
	}
}

func NewCtxChangeMessage(caseNumber string) model.Message {
	transactionID := uuid.New().String()

	return model.Message{
		Kind:          model.ContextChangeRequest,
		TransactionID: &transactionID,
		Context: []model.ContextItem{
			{Key: model.CaseNumber, Value: caseNumber},
		},
	}
}

func NewCtxAcceptMessage(transactionID string) model.Message {
	return model.Message{
		Kind:          model.ContextChangeAccept,
		TransactionID: &transactionID,
	}
}

func NewCtxRejectMessage(transactionID, reason string, status model.StatusCode) model.Message {
	rejection := model.MessageRejection{
		Reason: reason,
		Status: status,
	}

	return model.Message{
		Kind:          model.ContextChangeReject,
		TransactionID: &transactionID,
		Rejection:     &rejection,
	}
}

func ContextFromCaseNumber(caseNumber string) []model.ContextItem {
	return []model.ContextItem{
		{Key: "case", Value: caseNumber},
	}
}
