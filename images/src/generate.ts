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

    const client = new gptscript.GPTScript()
    const dataset = await client.createDataset(
      process.env.GPTSCRIPT_WORKSPACE_ID!,
      createHash('sha256').update(prompt).digest('hex').substring(0, 16),
      `Generated images for "${prompt}"`
    )

    // Download all images concurrently
    const imageUrls = response.data.map(image => image.url).filter(url => url != null)
    await Promise.all(
      imageUrls.map(url => download(client, dataset.id, url))
    )
    console.log(`Created dataset with ID ${dataset.id} with ${imageUrls.length} images`)
  } catch (error) {
    console.log('Error while generating images:', error);
    process.exit(1);
  }
}

async function download(client: gptscript.GPTScript, datasetId: string, imageUrl: string) {
  const response = await axios.get(imageUrl, {
    responseType: 'arraybuffer'
  })
  const content = Buffer.from(response.data, 'binary')

  await client.addDatasetElement(
    process.env.GPTSCRIPT_WORKSPACE_ID!,
    datasetId,
    createHash('sha256').update(imageUrl).digest('hex').substring(0, 8),
    '',
    content
  )
}

export { generateImages };
