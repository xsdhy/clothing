package llm

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	volcModel "github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
)

//文档:https://www.volcengine.com/docs/82379/1824121

func GenerateImagesByVolcengineProtocol(ctx context.Context, apiKey, model, prompt string, base64Images []string) (imageDataURL, assistantText string, err error) {
	client := arkruntime.NewClientWithApiKey(apiKey)

	var sequentialImageGeneration volcModel.SequentialImageGeneration = "disabled" //auto\disabled
	maxImages := 5
	generateReq := volcModel.GenerateImagesRequest{
		Model:                     model, //doubao-seedream-4-0-250828
		Prompt:                    prompt,
		Image:                     base64Images,
		Size:                      volcengine.String("4K"),                                      //可选值类型1：1K、2K、4K；类型2:默认值：2048x2048 总像素取值范围：[1280x720, 4096x4096]  宽高比取值范围：[1/16, 16] 推荐的宽高像素值：  宽高比 宽高像素值 1:1 2048x2048 4:3 2304x1728 3:4 1728x2304 16:9 2560x1440 9:16 1440x2560 3:2 2496x1664 2:3 1664x2496 21:9 3024x1296
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
		return
	}
	defer stream.Close()
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
				imageDataURL = *recv.Url
				fmt.Printf("recv.Size: %s, recv.Url: %s", recv.Size, *recv.Url)
			}
		}
		if recv.Type == "image_generation.completed" {
			if recv.Error == nil {
				fmt.Printf("recv.Usage: %v", *recv.Usage)
			}
		}
	}
	return imageDataURL, assistantText, nil
}
