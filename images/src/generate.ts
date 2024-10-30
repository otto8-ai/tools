import OpenAI from 'openai'
import * as gptscript from '@gptscript-ai/gptscript'
import axios from 'axios'
import { createHash } from 'node:crypto'

type ImageSize = '1024x1024' | '256x256' | '512x512' | '1792x1024' | '1024x1792';
type ImageQuality = 'standard' | 'hd';

const generateImages = async (
  model: string = 'dall-e-3',
  prompt: string = '',
  size: string = '1024x1024',
  quality: string = 'standard',
  quantity: number = 1
): Promise<void> => {
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
      model,
      prompt,
      size: size as ImageSize,
      quality: quality as ImageQuality,
      n: quantity,
    });

    // Download all images concurrently
    const imageUrls = response.data.map(image => image.url).filter(url => url != null)
    const client = new gptscript.GPTScript()
    const filePaths = await Promise.all(
      imageUrls.map(url => download(client, url))
    );

    // Output the workspace file paths of the generated images
    filePaths.forEach(filePath => {
      console.log(filePath)
    })
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
  let filePath= `${createHash('sha256').update(imageUrl).digest('hex')}.png`;
  if (process.env.OTTO_THREAD_ID !== undefined) {
    filePath = `files/${filePath}`
  }

  await client.writeFileInWorkspace(filePath, content)

  return filePath
}

export { generateImages };
