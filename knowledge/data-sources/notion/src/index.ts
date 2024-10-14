import { Client } from "@notionhq/client";
import dotenv from "dotenv";
import { mkdir, writeFile } from "fs/promises";
import path from "path";
import { PageObjectResponse } from "@notionhq/client/build/src/api-endpoints";
import { getPageContent } from "./page";
import * as fs from "node:fs";

dotenv.config();

interface Metadata {
  input: InputMetadata;
  output: OutputMetadata;
  outputDir: string;
}

interface InputMetadata {
  notionConfig: {
    pages: string[];
  };
  exclude: string[];
}

interface OutputMetadata {
  files: {
    [pageId: string]: {
      updatedAt: string;
      filePath: string;
      url: string;
    };
  };
  status: string;
  error: string;
  state: {
    notionState: {
      pages: Record<
        string,
        {
          url: string;
          title: string;
          folderPath: string;
        }
      >;
    };
  };
}

// Function to write a page to a file
async function writePageToFile(
  client: Client,
  page: PageObjectResponse,
  directory: string
) {
  const pageId = page.id;
  const pageContent = await getPageContent(client, pageId);
  const fileDir = path.join(directory, pageId.toString());
  await mkdir(fileDir, { recursive: true });
  const filePath = getPath(directory, page);
  fs.writeFileSync(filePath, pageContent, "utf8");
}

function getPath(directory: string, page: PageObjectResponse): string {
  const pageId = page.id;
  const fileDir = path.join(directory, pageId.toString());
  let title = (
    (page.properties?.title ?? page.properties?.Name) as any
  )?.title[0]?.plain_text
    ?.trim()
    .replaceAll(/\//g, "-");
  if (!title) {
    title = pageId.toString();
  }
  return path.join(fileDir, title + ".md");
}

function getTitle(page: PageObjectResponse): string {
  let title = (
    (page.properties?.title ?? page.properties?.Name) as any
  )?.title[0]?.plain_text
    ?.trim()
    .replaceAll(/\//g, "-");
  if (!title) {
    title = page.id.toString();
  }
  return title;
}

async function getPage(client: Client, pageId: string) {
  const page = await client.pages.retrieve({ page_id: pageId });
  return page as PageObjectResponse;
}

async function main() {
  const client = new Client({
    auth: process.env.NOTION_TOKEN,
  });
  let workingDir = process.env.GPTSCRIPT_WORKSPACE_DIR ?? process.cwd();

  // Fetch all pages
  let metadata: Metadata = {} as Metadata;
  const metadataPath = path.join(workingDir, ".metadata.json");
  if (fs.existsSync(metadataPath)) {
    metadata = JSON.parse(fs.readFileSync(metadataPath, "utf8").toString());
  }
  if (metadata.outputDir) {
    workingDir = metadata.outputDir;
  }
  console.log("Working directory:", workingDir);
  await mkdir(workingDir, { recursive: true });

  if (!metadata.output) {
    metadata.output = {} as OutputMetadata;
  }

  if (!metadata.output.files) {
    metadata.output.files = {};
  }

  if (!metadata.output.state) {
    metadata.output.state = {} as {
      notionState: {
        pages: Record<
          string,
          { url: string; title: string; folderPath: string }
        >;
      };
    };
  }

  if (!metadata.output.state.notionState) {
    metadata.output.state.notionState = {} as {
      pages: Record<string, { url: string; title: string; folderPath: string }>;
    };
  }

  if (!metadata.output.state.notionState.pages) {
    metadata.output.state.notionState.pages = {};
  }

  let syncedCount = 0;
  let error: any;
  try {
    const allPages = await client.search({
      filter: { property: "object", value: "page" },
    });

    for (const page of allPages.results) {
      let p = page as PageObjectResponse;
      if (p.archived) {
        continue;
      }
      const pageId = p.id;
      const pageUrl = p.url;
      const pageTitle = getTitle(p);

      let folderPath = "";
      while (p.parent && p.parent.type === "page_id") {
        try {
          const parentPage = await getPage(client, p.parent.page_id);
          const parentTitle = getTitle(parentPage);
          folderPath = path.join(parentTitle, folderPath);
          p = parentPage;
        } catch (err: any) {
          folderPath = "";
          break;
        }
      }
      metadata.output.state.notionState.pages[pageId] = {
        url: pageUrl,
        title: pageTitle,
        folderPath: folderPath,
      };
    }
    for (const pageId of Object.keys(metadata.output.state.notionState.pages)) {
      if (
        !allPages.results
          .filter((p) => !(p as PageObjectResponse).archived)
          .some((page) => (page as PageObjectResponse).id === pageId)
      ) {
        delete metadata.output.state.notionState.pages[pageId];
      }
    }

    if (metadata.input?.notionConfig?.pages) {
      for (const pageId of metadata.input.notionConfig.pages) {
        if (metadata.input.exclude?.includes(pageId)) {
          continue;
        }
        const page = await getPage(client, pageId);
        if (
          !metadata.output.files[pageId] ||
          metadata.output.files[pageId].updatedAt !== page.last_edited_time
        ) {
          await writePageToFile(client, page, workingDir);
          syncedCount++;
          metadata.output.files[pageId] = {
            url: page.url,
            filePath: getPath(workingDir, page!),
            updatedAt: page.last_edited_time,
          };
        }
        metadata.output.status = `${syncedCount}/${
          Object.keys(metadata.input.notionConfig.pages).length
        } number of pages have been synced`;
        await writeFile(metadataPath, JSON.stringify(metadata, null, 2));
      }
    }
    for (const [pageId, fileInfo] of Object.entries(metadata.output.files)) {
      if (
        !metadata.input?.notionConfig?.pages?.includes(pageId) ||
        metadata.input?.exclude?.includes(pageId)
      ) {
        try {
          await fs.rmSync(path.dirname(fileInfo.filePath), {
            recursive: true,
            force: true,
          });
          delete metadata.output.files[pageId];
          console.log(`Deleted file and entry for page ID: ${pageId}`);
        } catch (error) {
          console.error(`Failed to delete file ${fileInfo.filePath}:`, error);
        }
      }
    }
  } catch (err: any) {
    error = err;
    throw err;
  } finally {
    metadata.output.error = error?.message ?? "";
    metadata.output.status = ``;
    await writeFile(metadataPath, JSON.stringify(metadata, null, 2));
  }
}

main()
  .then(() => process.exit(0))
  .catch((err) => {
    console.error(err);
    process.exit(1);
  });
