package util

import (
	"tcs/internal/model"
)

func NewSubRequestMessage(application string, replaceExistingClient bool) model.Message {
	info := model.ConnectionInfo{
		Version:              1,
		Application:          application,
		ReplaceExitingClient: &replaceExistingClient,
	}

	return model.Message{
		Kind: model.SyncRequest,
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
		Kind:    model.SyncAccept,
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
		Kind:      model.SyncReject,
		Info:      &info,
		Rejection: &rejection,
	}
}

func NewCtxChangeMessage(caseNumber string) model.Message {
	return model.Message{
		Kind: model.ContextChangeRequest,
		Context: []model.ContextItem{
			{Key: model.CaseNumber, Value: caseNumber},
		},
	}
}

func NewCtxAcceptMessage(context []model.ContextItem) model.Message {
	return model.Message{
		Kind:    model.ContextChangeAccept,
		Context: context,
	}
}

func NewCtxRejectMessage(
	currentContext []model.ContextItem,
	rejectedContext []model.ContextItem,
	reason string,
	status model.StatusCode,
) model.Message {
	rejection := model.MessageRejection{
		Reason: reason,
		Status: status,
	}

	return model.Message{
		Kind:           model.ContextChangeReject,
		Context:        rejectedContext,
		CurrentContext: currentContext,
		Rejection:      &rejection,
	}
}

func ContextFromCaseNumber(caseNumber string) []model.ContextItem {
	return []model.ContextItem{
		{Key: "case", Value: caseNumber},
	}
}
