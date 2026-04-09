package cron

import (
	"context"
	"ucode/ucode_go_object_builder_service/genproto/company_service"
	"ucode/ucode_go_object_builder_service/grpc/client"
	"ucode/ucode_go_object_builder_service/pkg/logger"
	"ucode/ucode_go_object_builder_service/storage"

	pb "ucode/ucode_go_object_builder_service/genproto/company_service"

	"github.com/robfig/cron/v3"
)

type TaskScheduler struct {
	cronJob *cron.Cron
	logger  logger.LoggerI
	storage storage.StorageI
	svcs    client.ServiceManagerI
}

type TaskSchedulerI interface {
	RunJobs(context.Context) error
	DeleteFunctionLogs(context.Context) error
}

func New(log logger.LoggerI, storage storage.StorageI, svcs client.ServiceManagerI) TaskSchedulerI {
	var cronJob = cron.New()
	defer cronJob.Start()
	return &TaskScheduler{
		cronJob: cronJob,
		logger:  log,
		storage: storage,
		svcs:    svcs,
	}
}

func (t *TaskScheduler) RunJobs(ctx context.Context) error {
	t.logger.Info("Jobs Started:")

	_, err := t.cronJob.AddFunc("0 0 * * *", func() {
		err := t.DeleteFunctionLogs(ctx)
		if err != nil {
			t.logger.Error("error in DeleteFunctionLogs", logger.Error(err))
		}
	})

	return err
}

func (t *TaskScheduler) DeleteFunctionLogs(ctx context.Context) error {
	t.logger.Info("Running DeleteFunctionLogs job ...")

	// TODO pagination qilib ishlidgan qilish kere, bitta kop daniyla kelib qolishi mumkun

	response, err := t.svcs.ResourceService().GetListResourceEnvironment(ctx, &company_service.GetListResourceEnvironmentReq{
		ResourceType: pb.ResourceType_POSTGRESQL,
	})
	if err != nil {
		t.logger.Info("error in getting resource environment", logger.Error(err))
		return err
	}

	for i := range response.Data {
		err = t.storage.VersionHistory().DeleteFunctionLogs(ctx, response.Data[i].Id)
		if err != nil {
			t.logger.Info("error in deleting function logs", logger.Error(err))
			continue
		}
	}

	return nil
}
