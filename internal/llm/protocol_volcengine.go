package llm

import (
	"clothing/internal/entity"
	"clothing/internal/utils"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	volcModel "github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
)

//文档:https://www.volcengine.com/docs/82379/1824121

func GenerateContentByVolcengineProtocol(ctx context.Context, apiKey, model, prompt, size string, base64Images []string) (*entity.GenerateContentResponse, error) {
	client := arkruntime.NewClientWithApiKey(apiKey)

	var sequentialImageGeneration volcModel.SequentialImageGeneration = "auto" // allow multi-image sequences when supported
	maxImages := 5
	sizeValue := strings.TrimSpace(size)
	if sizeValue == "" {
		sizeValue = "4K"
	}
	generateReq := volcModel.GenerateImagesRequest{
		Model:                     model, //doubao-seedream-4-0-250828
		Prompt:                    prompt,
		Image:                     base64Images,
		Size:                      volcengine.String(sizeValue),                                 //可选值类型1：1K、2K、4K；类型2:默认值：2048x2048 总像素取值范围：[1280x720, 4096x4096]  宽高比取值范围：[1/16, 16] 推荐的宽高像素值：  宽高比 宽高像素值 1:1 2048x2048 4:3 2304x1728 3:4 1728x2304 16:9 2560x1440 9:16 1440x2560 3:2 2496x1664 2:3 1664x2496 21:9 3024x1296
		ResponseFormat:            volcengine.String(volcModel.GenerateImagesResponseFormatURL), //指定生成图像的返回格式：url：返回图片下载链接；链接在图片生成后24小时内有效。b64_json：以 Base64 编码字符串的 JSON 格式返回图像数据。
		Watermark:                 volcengine.Bool(false),                                       //是否在生成的图片中添加水印。
		SequentialImageGeneration: &sequentialImageGeneration,                                   //控制是否关闭组图功能（基于您输入的内容，生成的一组内容关联的图片）auto：自动判断模式，模型会根据用户提供的提示词自主判断是否返回组图以及组图包含的图片数量。disabled：关闭组图功能，模型只会生成一张图。
		SequentialImageGenerationOptions: &volcModel.SequentialImageGenerationOptions{
			MaxImages: &maxImages,
		}, //指定本次请求，最多可生成的图片数量。仅当sequential_image_generation为auto时生效。
	}
	stream, err := client.GenerateImagesStreaming(ctx, generateReq)
	if err != nil {
		fmt.Printf("call GenerateImagesStreaming error: %v", err)
		return nil, err
	}
	defer stream.Close()

	var imageDataURLs []string
	var assistantText string

	for {
		recv, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			assistantText = recv.Error.Message
			fmt.Printf("Stream generate images error: %v", err)
			break
		}
		if recv.Type == "image_generation.partial_failed" {
			fmt.Printf("Stream generate images error: %v", recv.Error)
			assistantText = recv.Error.Message
			if strings.EqualFold(recv.Error.Code, "InternalServiceError") {
				break
			}
		}
		if recv.Type == "image_generation.partial_succeeded" {
			if recv.Error == nil && recv.Url != nil {
				url := strings.TrimSpace(*recv.Url)
				if url != "" {
					imageDataURLs = append(imageDataURLs, url)
					logrus.WithFields(logrus.Fields{
						"url":  recv.Url,
						"site": recv.Size,
					}).Info("image data url")
				}
			}
		}
		if recv.Type == "image_generation.completed" {
			logrus.WithFields(logrus.Fields{
				"recvType": recv.Type,
				"recvErr":  recv.Error,
			}).Info("image generation complete")
			break
		}
	}
	if len(imageDataURLs) == 0 {
		return &entity.GenerateContentResponse{
			TextContent: assistantText,
		}, errors.New(assistantText)
	}
	return &entity.GenerateContentResponse{
		ImageAssets: imageDataURLs,
		TextContent: assistantText,
	}, nil
}

func GenerateVolcengineVideo(ctx context.Context, apiKey string, model entity.DbModel, prompt, size string, duration int, images []string) (*entity.GenerateContentResponse, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.New("api key missing")
	}

	client := arkruntime.NewClientWithApiKey(apiKey)
	trimmedPrompt := buildVolcengineVideoPrompt(prompt, size, duration)

	contentItems := make([]*volcModel.CreateContentGenerationContentItem, 0, len(images)+1)
	if strings.TrimSpace(trimmedPrompt) != "" {
		contentItems = append(contentItems, &volcModel.CreateContentGenerationContentItem{
			Type: volcModel.ContentGenerationContentItemTypeText,
			Text: volcengine.String(trimmedPrompt),
		})
	}

	imageItems, err := buildVolcengineVideoImages(ctx, model.ModelID, images)
	if err != nil {
		return nil, err
	}
	contentItems = append(contentItems, imageItems...)

	if len(contentItems) == 0 {
		return nil, errors.New("volcengine video content is empty")
	}

	logrus.WithFields(logrus.Fields{
		"model":         model.ModelID,
		"prompt_length": len(trimmedPrompt),
		"image_count":   len(images),
	}).Info("volcengine_generate_video_start")

	req := volcModel.CreateContentGenerationTaskRequest{
		Model:   model.ModelID,
		Content: contentItems,
	}

	createResp, err := client.CreateContentGenerationTask(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("volcengine create video task: %w", err)
	}

	taskID := strings.TrimSpace(createResp.ID)
	requestID := extractVolcengineRequestID(createResp.Header())
	if taskID == "" {
		return &entity.GenerateContentResponse{RequestID: requestID}, errors.New("volcengine video task id is empty")
	}

	assets, revisedPrompt, err := waitForVolcengineVideo(ctx, client, taskID)
	if err != nil {
		return &entity.GenerateContentResponse{
			TextContent:      revisedPrompt,
			ExternalTaskCode: taskID,
			RequestID:        requestID,
		}, err
	}
	if len(assets) == 0 {
		return &entity.GenerateContentResponse{
			TextContent:      revisedPrompt,
			ExternalTaskCode: taskID,
			RequestID:        requestID,
		}, errors.New("volcengine video response missing video url")
	}

	return &entity.GenerateContentResponse{
		ImageAssets:      assets,
		TextContent:      revisedPrompt,
		ExternalTaskCode: taskID,
		RequestID:        requestID,
	}, nil
}

func buildVolcengineVideoImages(ctx context.Context, modelID string, images []string) ([]*volcModel.CreateContentGenerationContentItem, error) {
	trimmed := make([]string, 0, len(images))
	for _, img := range images {
		img = strings.TrimSpace(img)
		if img == "" {
			continue
		}
		trimmed = append(trimmed, img)
	}
	if len(trimmed) == 0 {
		return nil, nil
	}

	modelLower := strings.ToLower(modelID)
	useReference := len(trimmed) > 2 && strings.Contains(modelLower, "lite") && strings.Contains(modelLower, "i2v")

	if useReference {
		maxRefs := 4
		if len(trimmed) > maxRefs {
			logrus.WithFields(logrus.Fields{
				"model":       modelID,
				"image_count": len(trimmed),
				"max_refs":    maxRefs,
			}).Info("volcengine reference images trimmed")
		}

		items := make([]*volcModel.CreateContentGenerationContentItem, 0, maxRefs)
		for idx, img := range trimmed {
			if idx >= maxRefs {
				break
			}
			url, err := inlineVolcengineImage(ctx, img)
			if err != nil {
				return nil, fmt.Errorf("prepare reference image: %w", err)
			}
			items = append(items, &volcModel.CreateContentGenerationContentItem{
				Type: volcModel.ContentGenerationContentItemTypeImage,
				ImageURL: &volcModel.ImageURL{
					URL: url,
				},
				Role: volcengine.String("reference_image"),
			})
		}
		return items, nil
	}

	if len(trimmed) >= 2 {
		if len(trimmed) > 2 {
			logrus.WithFields(logrus.Fields{
				"model":       modelID,
				"image_count": len(trimmed),
			}).Info("volcengine using first and last frame")
		}

		firstURL, err := inlineVolcengineImage(ctx, trimmed[0])
		if err != nil {
			return nil, fmt.Errorf("prepare first frame: %w", err)
		}
		lastURL, err := inlineVolcengineImage(ctx, trimmed[len(trimmed)-1])
		if err != nil {
			return nil, fmt.Errorf("prepare last frame: %w", err)
		}

		return []*volcModel.CreateContentGenerationContentItem{
			{
				Type: volcModel.ContentGenerationContentItemTypeImage,
				ImageURL: &volcModel.ImageURL{
					URL: firstURL,
				},
				Role: volcengine.String("first_frame"),
			},
			{
				Type: volcModel.ContentGenerationContentItemTypeImage,
				ImageURL: &volcModel.ImageURL{
					URL: lastURL,
				},
				Role: volcengine.String("last_frame"),
			},
		}, nil
	}

	soloURL, err := inlineVolcengineImage(ctx, trimmed[0])
	if err != nil {
		return nil, fmt.Errorf("prepare first frame: %w", err)
	}

	return []*volcModel.CreateContentGenerationContentItem{
		{
			Type: volcModel.ContentGenerationContentItemTypeImage,
			ImageURL: &volcModel.ImageURL{
				URL: soloURL,
			},
			Role: volcengine.String("first_frame"),
		},
	}, nil
}

func buildVolcengineVideoPrompt(prompt, size string, duration int) string {
	trimmed := strings.TrimSpace(prompt)
	if trimmed == "" {
		return trimmed
	}

	promptLower := strings.ToLower(trimmed)
	if sizeValue := strings.TrimSpace(size); sizeValue != "" && !strings.Contains(promptLower, "--rs") {
		trimmed += fmt.Sprintf(" --rs %s", strings.ToLower(sizeValue))
	}
	if duration > 0 && !strings.Contains(promptLower, "--dur") {
		trimmed += fmt.Sprintf(" --dur %d", duration)
	}
	return trimmed
}

func inlineVolcengineImage(ctx context.Context, payload string) (string, error) {
	trimmed := strings.TrimSpace(payload)
	if trimmed == "" {
		return "", errors.New("empty image payload")
	}
	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		b64, mimeType, err := downloadImageAsBase64(ctx, trimmed)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("data:%s;base64,%s", fallbackMime(mimeType), b64), nil
	}
	return utils.EnsureDataURL(trimmed), nil
}

func waitForVolcengineVideo(ctx context.Context, client *arkruntime.Client, taskID string) ([]string, string, error) {
	if strings.TrimSpace(taskID) == "" {
		return nil, "", errors.New("volcengine missing task id for video")
	}

	pollInterval := 5 * time.Second

	for {
		if ctx.Err() != nil {
			return nil, "", ctx.Err()
		}

		resp, err := client.GetContentGenerationTask(ctx, volcModel.GetContentGenerationTaskRequest{ID: taskID})
		if err != nil {
			return nil, "", fmt.Errorf("volcengine get video task: %w", err)
		}

		status := strings.ToLower(strings.TrimSpace(resp.Status))
		revisedPrompt := ""
		if resp.RevisedPrompt != nil {
			revisedPrompt = strings.TrimSpace(*resp.RevisedPrompt)
		}

		if status == strings.ToLower(volcModel.StatusSucceeded) {
			return collectVolcengineVideoAssets(resp.Content), revisedPrompt, nil
		}

		if resp.Error != nil && resp.Error.Message != "" {
			return nil, revisedPrompt, fmt.Errorf("volcengine task error: %s", resp.Error.Message)
		}

		if status == strings.ToLower(volcModel.StatusFailed) ||
			status == strings.ToLower(volcModel.StatusCancelled) ||
			status == "expired" {
			return nil, revisedPrompt, fmt.Errorf("volcengine task %s", status)
		}

		logrus.WithFields(logrus.Fields{
			"task_id": taskID,
			"status":  status,
		}).Info("volcengine video task still running")

		select {
		case <-ctx.Done():
			return nil, revisedPrompt, ctx.Err()
		case <-time.After(pollInterval):
		}
	}
}

func collectVolcengineVideoAssets(content volcModel.Content) []string {
	assets := make([]string, 0, 2)
	if url := strings.TrimSpace(content.VideoURL); url != "" {
		assets = append(assets, url)
	}
	if url := strings.TrimSpace(content.LastFrameURL); url != "" {
		assets = append(assets, url)
	}
	return assets
}

func extractVolcengineRequestID(header http.Header) string {
	reqID := strings.TrimSpace(header.Get("X-Request-Id"))
	if reqID != "" {
		return reqID
	}
	return strings.TrimSpace(header.Get(volcModel.ClientRequestHeader))
}
