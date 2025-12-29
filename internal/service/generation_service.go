package service

import (
	"clothing/internal/entity"
	"clothing/internal/llm"
	"clothing/internal/model"
	"clothing/internal/storage"
	"clothing/internal/utils"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// GenerationService 内容生成服务，封装生成相关的业务逻辑
type GenerationService struct {
	repo    model.Repository
	storage storage.Storage

	// notifyFunc 用于通知生成完成事件（由调用方设置）
	notifyFunc func(clientID string, recordID uint, status string, errMsg string)
}

// NewGenerationService 创建生成服务实例
func NewGenerationService(repo model.Repository, store storage.Storage) *GenerationService {
	return &GenerationService{
		repo:    repo,
		storage: store,
	}
}

// SetNotifyFunc 设置通知函数（用于 SSE 推送）
func (s *GenerationService) SetNotifyFunc(fn func(clientID string, recordID uint, status string, errMsg string)) {
	s.notifyFunc = fn
}

// GenerateContentRequest 生成内容请求参数
type GenerateContentRequest struct {
	Record   entity.DbUsageRecord
	Request  entity.GenerateContentRequest
	Model    entity.DbModel
	Service  llm.AIService
	ClientID string
}

// GenerateContentAsync 异步生成内容
func (s *GenerationService) GenerateContentAsync(req GenerateContentRequest) {
	go s.handleGeneration(req)
}

// handleGeneration 处理内容生成的核心逻辑
func (s *GenerationService) handleGeneration(req GenerateContentRequest) {
	if s.repo == nil {
		return
	}

	record := req.Record
	request := req.Request
	dbModel := req.Model
	service := req.Service
	clientID := strings.TrimSpace(req.ClientID)

	genCtx, cancelGen := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancelGen()

	var updates entity.UsageRecordUpdates
	completionError := ""
	var storageIssues []string

	// 保存输入图片
	if len(request.Inputs.Images) > 0 {
		inputPaths, err := s.saveMediaToStorage(genCtx, "inputs", request.Inputs.Images, request.ModelID)
		if len(inputPaths) > 0 {
			inputImages := entity.StringArray(inputPaths)
			updates.InputImages = &inputImages
		}
		if err != nil {
			storageIssues = append(storageIssues, fmt.Sprintf("input images: %v", err))
			logrus.WithError(err).WithFields(logrus.Fields{
				"record_id": record.ID,
				"provider":  record.ProviderID,
				"model":     record.ModelID,
			}).Warn("failed to persist input images")
		}
	}

	// 调用 LLM 服务生成内容
	resp, err := service.GenerateContent(genCtx, request, dbModel)

	var externalTaskCode, requestID string
	var outputs []string
	var text string

	if resp != nil {
		externalTaskCode = resp.ExternalTaskCode
		requestID = resp.RequestID
		outputs = resp.ImageAssets
		text = resp.TextContent
	}

	if externalTaskCode != "" {
		updates.ExternalTaskCode = &externalTaskCode
	}
	if requestID != "" {
		updates.RequestID = &requestID
	}

	// 处理生成失败的情况
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"record_id": record.ID,
			"provider":  record.ProviderID,
			"model":     record.ModelID,
		}).Error("failed to generate content")

		errMsg := err.Error()
		if len(storageIssues) > 0 {
			errMsg = appendStorageNotes(errMsg, storageIssues)
		}

		updates.ErrorMessage = &errMsg
		s.updateUsageRecord(record.ID, updates)
		s.notifyComplete(clientID, record.ID, "failure", errMsg)
		return
	}

	logrus.WithFields(logrus.Fields{
		"record_id": record.ID,
		"provider":  record.ProviderID,
		"model":     record.ModelID,
	}).Info("generated content")

	// 保存文本内容
	if text != "" {
		updates.TextContent = &text
	}

	// 保存输出媒体文件
	if len(outputs) > 0 {
		outputPaths, err := s.saveMediaToStorage(genCtx, "outputs", outputs, request.ModelID)
		if len(outputPaths) > 0 {
			outputImages := entity.StringArray(outputPaths)
			updates.OutputImages = &outputImages
		}
		if err != nil {
			storageIssues = append(storageIssues, fmt.Sprintf("output assets: %v", err))
			logrus.WithError(err).WithFields(logrus.Fields{
				"record_id": record.ID,
				"provider":  record.ProviderID,
				"model":     record.ModelID,
			}).Warn("failed to persist output assets")
		}
	}

	// 合并存储问题到错误信息
	if len(storageIssues) > 0 {
		existingError := ""
		if updates.ErrorMessage != nil {
			existingError = *updates.ErrorMessage
		}
		combined := appendStorageNotes(existingError, storageIssues)
		updates.ErrorMessage = &combined
		completionError = combined
	}

	s.updateUsageRecord(record.ID, updates)
	s.notifyComplete(clientID, record.ID, "success", completionError)
}

// saveMediaToStorage 保存媒体文件到存储
func (s *GenerationService) saveMediaToStorage(parentCtx context.Context, category string, payloads []string, modelName string) ([]string, error) {
	if s.storage == nil || len(payloads) == 0 {
		return nil, nil
	}

	if parentCtx == nil {
		parentCtx = context.Background()
	}

	ctx, cancel := context.WithTimeout(parentCtx, 5*time.Minute)
	defer cancel()

	var (
		paths []string
		errs  []string
	)

	for idx, payload := range payloads {
		trimmed := strings.TrimSpace(payload)
		if trimmed == "" {
			continue
		}

		data, ext, err := s.resolveMediaPayload(ctx, trimmed)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%d: %v", idx, err))
			continue
		}

		saveOpts := storage.SaveOptions{Category: category, Extension: ext}
		switch strings.ToLower(strings.TrimSpace(category)) {
		case "inputs":
			saveOpts.SkipIfExists = true
			saveOpts.BaseName = computeInputBaseName(data)
		case "outputs":
			saveOpts.BaseName = buildOutputBaseName(modelName, idx)
		default:
			saveOpts.BaseName = ""
		}

		relPath, err := s.storage.Save(ctx, data, saveOpts)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%d: %v", idx, err))
			continue
		}
		paths = append(paths, relPath)
	}

	if len(errs) > 0 {
		return paths, fmt.Errorf("%s", strings.Join(errs, "; "))
	}

	return paths, nil
}

// resolveMediaPayload 解析媒体数据（URL 或 base64）
func (s *GenerationService) resolveMediaPayload(ctx context.Context, payload string) ([]byte, string, error) {
	trimmed := strings.TrimSpace(payload)
	if trimmed == "" {
		return nil, "", fmt.Errorf("empty payload")
	}

	// 处理 URL
	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		reqCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, trimmed, nil)
		if err != nil {
			return nil, "", fmt.Errorf("create request: %w", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, "", fmt.Errorf("download image: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, "", fmt.Errorf("download image http %d", resp.StatusCode)
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, "", fmt.Errorf("read image body: %w", err)
		}

		ext := utils.ExtensionFromMime(resp.Header.Get("Content-Type"))
		if ext == "" {
			ext = utils.ExtensionFromMime(http.DetectContentType(data))
		}
		if ext == "" {
			ext = "bin"
		}

		return data, ext, nil
	}

	// 处理 base64
	data, ext, err := utils.DecodeMediaPayload(trimmed)
	if err == nil {
		return data, ext, nil
	}

	data, ext, err = utils.DecodeMediaPayload(utils.EnsureDataURL(trimmed))
	if err != nil {
		return nil, "", err
	}

	return data, ext, nil
}

// updateUsageRecord 更新使用记录
func (s *GenerationService) updateUsageRecord(recordID uint, updates entity.UsageRecordUpdates) {
	if s.repo == nil || recordID == 0 || updates.IsEmpty() {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.repo.UpdateUsageRecord(ctx, recordID, updates); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"record_id": recordID,
		}).Error("failed to update usage record")
	}
}

// notifyComplete 通知生成完成
func (s *GenerationService) notifyComplete(clientID string, recordID uint, status string, errMsg string) {
	if s.notifyFunc != nil && strings.TrimSpace(clientID) != "" {
		s.notifyFunc(clientID, recordID, status, errMsg)
	}
}

// appendStorageNotes 合并存储问题说明
func appendStorageNotes(existing string, notes []string) string {
	if len(notes) == 0 {
		return existing
	}
	combined := strings.Join(notes, "; ")
	if strings.TrimSpace(existing) == "" {
		return combined
	}
	return existing + "; " + combined
}

// computeInputBaseName 计算输入文件的基础名称（使用 MD5 哈希）
func computeInputBaseName(data []byte) string {
	sum := md5.Sum(data)
	return hex.EncodeToString(sum[:])
}

// buildOutputBaseName 构建输出文件的基础名称
func buildOutputBaseName(modelName string, idx int) string {
	token := storage.SanitizeToken(modelName)
	if token == "" {
		token = "model"
	}
	if len(token) > 32 {
		token = token[:32]
	}
	suffix := time.Now().UTC().UnixNano()
	return fmt.Sprintf("%s_%d_%d", token, suffix, idx)
}
