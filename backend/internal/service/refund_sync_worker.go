package service

import (
	"errors"
	"time"

	"deskorder/config"
	"deskorder/internal/model"
	"deskorder/internal/repository"

	"github.com/rs/zerolog/log"
)

const (
	defaultRechargeRefundSyncInterval = 60 * time.Second
	rechargeRefundSyncBatchSize       = 50
	maxRechargeRefundAutoSyncRetries  = 8
	maxRechargeRefundAutoSyncBackoff  = 15 * time.Minute
)

func StartRechargeRefundSyncWorker() {
	interval := resolveRechargeRefundSyncInterval()
	go runRechargeRefundSyncWorker(interval)
	log.Info().Dur("interval", interval).Msg("recharge refund sync worker started")
}

func runRechargeRefundSyncWorker(interval time.Duration) {
	syncProcessingRechargeRefundRequests()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		syncProcessingRechargeRefundRequests()
	}
}

func syncProcessingRechargeRefundRequests() {
	list, total, err := repository.ListRefundRequests(1, rechargeRefundSyncBatchSize, model.RefundReviewScopeRecharge, model.RefundStatusProcessing, "")
	if err != nil {
		log.Error().Err(err).Msg("failed to list processing recharge refund requests")
		return
	}
	if len(list) == 0 {
		return
	}
	processingCount := 0
	successCount := 0
	failedCount := 0
	skippedCount := 0
	errorCount := 0
	for i := range list {
		execution, execErr := repository.GetLatestRefundExecutionByRequestID(list[i].ID)
		if execErr != nil {
			errorCount++
			log.Warn().Err(execErr).Uint("refund_request_id", list[i].ID).Msg("failed to load latest recharge refund execution")
			continue
		}
		if !shouldAutoSyncRechargeRefundExecution(execution, time.Now()) {
			skippedCount++
			continue
		}
		resp, syncErr := AdminSyncRechargeRefundRequestStatus(list[i].ID)
		if syncErr != nil {
			if isIgnorableRechargeRefundSyncError(syncErr) {
				continue
			}
			errorCount++
			log.Warn().Err(syncErr).Uint("refund_request_id", list[i].ID).Msg("failed to sync recharge refund request status")
			continue
		}
		switch resp.Status {
		case model.RefundStatusSuccess:
			successCount++
		case model.RefundStatusFailed:
			failedCount++
		case model.RefundStatusProcessing:
			processingCount++
		}
	}
	log.Info().
		Int64("queued_total", total).
		Int("scanned", len(list)).
		Int("skipped", skippedCount).
		Int("still_processing", processingCount).
		Int("success", successCount).
		Int("failed", failedCount).
		Int("errors", errorCount).
		Msg("recharge refund sync worker pass completed")
}

func resolveRechargeRefundSyncInterval() time.Duration {
	if config.AppConfig == nil || config.AppConfig.WeChat.RefundSyncIntervalSeconds <= 0 {
		return defaultRechargeRefundSyncInterval
	}
	return time.Duration(config.AppConfig.WeChat.RefundSyncIntervalSeconds) * time.Second
}

func isIgnorableRechargeRefundSyncError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, nil) || err.Error() == "当前退款申请状态不可同步" || err.Error() == "当前退款执行状态不可同步" || err.Error() == "退款申请不存在"
}

func shouldAutoSyncRechargeRefundExecution(item *model.RefundExecution, now time.Time) bool {
	if item == nil {
		return false
	}
	if isRechargeRefundAutoSyncPaused(item) {
		return false
	}
	next := calculateNextRechargeRefundAutoSyncAt(item)
	if next == nil {
		return true
	}
	return !now.Before(*next)
}

func isRechargeRefundAutoSyncPaused(item *model.RefundExecution) bool {
	if item == nil {
		return false
	}
	if item.Status != model.RefundExecutionStatusProcessing {
		return false
	}
	return item.RetryCount >= maxRechargeRefundAutoSyncRetries
}

func calculateNextRechargeRefundAutoSyncAt(item *model.RefundExecution) *time.Time {
	if item == nil || item.Status != model.RefundExecutionStatusProcessing || isRechargeRefundAutoSyncPaused(item) {
		return nil
	}
	backoff := calculateRechargeRefundAutoSyncBackoff(item.RetryCount)
	next := item.UpdatedAt.Add(backoff)
	return &next
}

func calculateRechargeRefundAutoSyncBackoff(retryCount int) time.Duration {
	base := resolveRechargeRefundSyncInterval()
	if base <= 0 {
		base = defaultRechargeRefundSyncInterval
	}
	if retryCount <= 0 {
		return base
	}
	backoff := base
	for step := 0; step < retryCount && backoff < maxRechargeRefundAutoSyncBackoff; step++ {
		backoff *= 2
		if backoff >= maxRechargeRefundAutoSyncBackoff {
			return maxRechargeRefundAutoSyncBackoff
		}
	}
	if backoff > maxRechargeRefundAutoSyncBackoff {
		return maxRechargeRefundAutoSyncBackoff
	}
	return backoff
}
