package web

import (
	"fmt"

	"github.com/fireeye/gocrack/server/storage"
	"github.com/fireeye/gocrack/shared"
)

func setEngineFile(stor storage.Backend, engineFileID string) *EngineFileItem {
	ef, err := stor.GetEngineFileByID(engineFileID)
	if err != nil || ef == nil {
		return nil
	}

	sef := convStorageEngineFile(*ef)
	return &sef
}

func convertStorageTaskToItem(stor storage.Backend, t storage.Task) TaskInfoResponseItem {
	item := TaskInfoResponseItem{
		TaskID:            t.TaskID,
		TaskName:          t.TaskName,
		CaseCode:          t.CaseCode,
		Comment:           t.Comment,
		AssignedToHost:    t.AssignedToHost,
		AssignedToDevices: t.AssignedToDevices,
		Status:            t.Status,
		CreatedBy:         t.CreatedBy,
		CreatedByUUID:     t.CreatedByUUID,
		CreatedAt:         t.CreatedAt,
		Engine:            TaskCrackEngineFancy(t.Engine),
		FileID:            t.FileID,
		Priority:          TaskPriorityFancy(t.Priority),
		Error:             t.Error,
	}

	switch ep := t.EnginePayload.(type) {
	case shared.HashcatUserOptions:
		hcitem := HashcatEnginePayload{}

		switch ep.AttackMode {
		case shared.AttackModeBruteForce:
			hcitem.AttackMode = "Brute Force"
			if ep.Masks != nil {
				hcitem.Masks = setEngineFile(stor, *ep.Masks)
			}
		case shared.AttackModeStraight:
			hcitem.AttackMode = "Straight"
			if ep.DictionaryFile != nil {
				hcitem.DictionaryFile = setEngineFile(stor, *ep.DictionaryFile)
			}

			if ep.ManglingRuleFile != nil {
				hcitem.ManglingRuleFile = setEngineFile(stor, *ep.ManglingRuleFile)
			}
		default:
			hcitem.AttackMode = "Unknown"
		}

		if htype := shared.LookupHashcatHashType(ep.HashType); htype != nil {
			hcitem.HashType = htype.Name
		} else {
			hcitem.HashType = fmt.Sprintf("Unknown HashType %d", ep.HashType)
		}
		item.EnginePayload = hcitem
	default:
		item.EnginePayload = ep
	}
	return item
}
