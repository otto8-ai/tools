import { fileTypeFromBuffer } from 'file-type';
import fs from 'fs';
import OpenAI from 'openai';
import { ChatCompletionContentPartImage } from 'openai/resources/chat/completions';

const resolveImageURL = async (image: string): Promise<string> => {
  const uri = new URL(image)
  switch (uri.protocol) {
    case 'http:':
    case 'https:':
      return image;
    case 'file:': {
      const filePath = image.slice(7);
      const data = fs.readFileSync(filePath);
      const mime = (await fileTypeFromBuffer(data))?.mime;
      if (mime !== 'image/jpeg' && mime !== 'image/png') {
        throw new Error('Unsupported mimetype');
      }
      const base64 = data.toString('base64')
      return `data:${mime};base64,${base64}`
    }
    default:
      throw new Error('Unsupported protocol')
  }
};

const analyzeImages = async (
  model: string = 'gpt-4o',
  prompt: string = '',
  images: string[] = [],
): Promise<void> => {
  if (!prompt) {
    throw new Error('No prompt provided. Please provide a prompt to send to the vision model.');
  }
  if (images.length === 0) {
    throw new Error('No images provided. Please provide a list of images to send to the vision model.');
  }

  const content = await Promise.all(
    images.map(async image => ({
      type: 'image_url',
      image_url: {
        detail: 'auto',
        url: await resolveImageURL(image),
      },
    }))
  ) as ChatCompletionContentPartImage[];

  const openai = new OpenAI();
  const response = await openai.chat.completions.create({
    model: model,
    stream: true,
    messages: [
      {
        role: 'user',
        content: [{ type: 'text', text: prompt }, ...content],
      },
    ],
  });

  for await (const part of response) {
    const { choices } = part;
    const text = choices[0]?.delta?.content;
    if (text) {
      process.stdout.write(text);
    }
  }
}

export { analyzeImages }
