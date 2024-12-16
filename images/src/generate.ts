import OpenAI from "openai"
import * as gptscript from "@gptscript-ai/gptscript"
import axios from "axios"
import {createHash} from "node:crypto"

type ImageSize = '1024x1024' | '256x256' | '512x512' | '1792x1024' | '1024x1792';
type ImageQuality = 'standard' | 'hd';

const threadId = process.env.ACORN_THREAD_ID
const acornServerUrl = process.env.ACORN_SERVER_URL
const downloadBaseUrl = (threadId && acornServerUrl) ? `${acornServerUrl}/api/threads/${threadId}/file` : null

export async function generateImages(
  prompt: string = '',
  size: string = '1024x1024',
  quality: string = 'standard',
  quantity: number = 1
): Promise<void> {
  if (!prompt) {
    throw new Error('No prompt provided. Please provide a prompt to generate images.');
  }

  if (!['1024x1024', '256x256', '512x512', '1792x1024', '1024x1792'].includes(size)) {
    throw new Error(`Invalid image size ${size}`)
  }

  if (!['standard', 'hd'].includes(quality)) {
    throw new Error(`Invalid image quality ${quality}`)
  }

  if (quantity < 1 || quantity > 10) {
    throw new Error(`Invalid image quantity ${quantity}`)
  }

  const openai = new OpenAI();

  try {
    const response = await openai.images.generate({
      model: process.env.ACORN_DEFAULT_IMAGE_GENERATION_MODEL ?? 'dall-e-3',
      prompt,
      size: size as ImageSize,
      quality: quality as ImageQuality,
      n: quantity,
    });

    // Download all images concurrently
    const imageUrls = response.data.map(image => image.url).filter(url => url != null)
    const client = new gptscript.GPTScript()
    const generatedImages= await Promise.all(
      imageUrls.map(async (url: string) => {
        const filePath = await download(client, url)
        if (!downloadBaseUrl) {
          return {
            filePath
          }
        }

        return {
          workspaceFilePath: filePath,
          downloadUrl: `${downloadBaseUrl}/${filePath}`
        }
      })
    );

    // Output the workspace file paths of the generated images
    console.log(JSON.stringify({
        prompt,
        images: generatedImages
      })
    )
  } catch (error) {
    console.log('Error while generating images:', error);
    process.exit(1);
  }
}

async function download(client: gptscript.GPTScript, imageUrl: string): Promise<string> {
  const response = await axios.get(imageUrl, {
    responseType: 'arraybuffer'
  })
  const content = Buffer.from(response.data, 'binary')

  // Generate a SHA-256 hash of the imageURL to use as the filename
  const filePath = `generated_image_${createHash('sha256').update(imageUrl).digest('hex').substring(0, 8)}.png`;

  await client.writeFileInWorkspace(`${threadId ? 'files/' : ''}${filePath}`, content);

  return filePath
}
