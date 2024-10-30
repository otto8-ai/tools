import { Client } from "@notionhq/client";
import dotenv from "dotenv";
import path from "path";
import { PageObjectResponse } from "@notionhq/client/build/src/api-endpoints";
import { getPageContent } from "./page";

dotenv.config();

interface OutputMetadata {
  files: {
    [pageId: string]: {
      updatedAt: string;
      filePath: string;
      url: string;
    };
  };
  status: string;
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

async function writePageToFile(
  client: Client,
  page: PageObjectResponse,
  gptscriptClient: any
) {
  const pageId = page.id;
  const pageContent = await getPageContent(client, pageId);
  const filePath = getPath(page);
  await gptscriptClient.writeFileInWorkspace(
    filePath,
    Buffer.from(pageContent)
  );
}

function getPath(page: PageObjectResponse): string {
  const pageId = page.id;
  const fileDir = path.join(pageId.toString());
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

  const gptscript = await import("@gptscript-ai/gptscript");
  const gptscriptClient = new gptscript.GPTScript();

  let output: OutputMetadata = {} as OutputMetadata;
  let metadataFile;
  try {
    metadataFile = await gptscriptClient.readFileInWorkspace(".metadata.json");
  } catch (err: any) {
    // Ignore any error if the metadata file doesn't exist. Ideally we should check for only not existing error but sdk doesn't provide that
  }
  if (metadataFile) {
    output = JSON.parse(metadataFile.toString());
  }

  if (!output.files) {
    output.files = {};
  }

  if (!output.state) {
    output.state = {} as {
      notionState: {
        pages: Record<
          string,
          { url: string; title: string; folderPath: string }
        >;
      };
    };
  }

  if (!output.state.notionState) {
    output.state.notionState = {} as {
      pages: Record<string, { url: string; title: string; folderPath: string }>;
    };
  }

  if (!output.state.notionState.pages) {
    output.state.notionState.pages = {};
  }

  let syncedCount = 0;
  const allPages = await client.search({
    filter: { property: "object", value: "page" },
  });

  const pageIds = new Set();
  for (const page of allPages.results) {
    let p = page as PageObjectResponse;
    if (p.archived) {
      continue;
    }
    const pageId = p.id;
    const pageUrl = p.url;
    const pageTitle = getTitle(p);
    pageIds.add(pageId);
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
    output.state.notionState.pages[pageUrl] = {
      url: pageUrl,
      title: pageTitle,
      folderPath: folderPath,
    };
  }

  await gptscriptClient.writeFileInWorkspace(
    ".metadata.json",
    Buffer.from(JSON.stringify(output, null, 2))
  );

  for (const pageId of Object.keys(output.state.notionState.pages)) {
    if (
      !allPages.results
        .filter((p) => !(p as PageObjectResponse).archived)
        .some((page) => (page as PageObjectResponse).id === pageId)
    ) {
      delete output.state.notionState.pages[pageId];
    }
  }

  for (const pageId of Object.keys(output.state.notionState.pages)) {
    const page = await getPage(client, pageId);
    if (
      !output.files[pageId] ||
      output.files[pageId].updatedAt !== page.last_edited_time
    ) {
      console.error(`Writing page url: ${page.url}`);
      await writePageToFile(client, page, gptscriptClient);
      output.files[pageId] = {
        url: page.url,
        filePath: getPath(page!),
        updatedAt: page.last_edited_time,
      };
    } else {
      console.error(`Skipping page url: ${page.url}`);
    }
    syncedCount++;
    output.status = `${syncedCount}/${
      Object.keys(output.state.notionState.pages).length
    } number of pages have been synced`;
    await gptscriptClient.writeFileInWorkspace(
      ".metadata.json",
      Buffer.from(JSON.stringify(output, null, 2))
    );
  }
  for (const [pageId, fileInfo] of Object.entries(output.files)) {
    if (!pageIds.has(pageId)) {
      try {
        await gptscriptClient.deleteFileInWorkspace(fileInfo.filePath);
        delete output.files[pageId];
        console.error(`Deleted file and entry for page ID: ${pageId}`);
      } catch (error) {
        console.error(`Failed to delete file ${fileInfo.filePath}:`, error);
      }
    }
  }
  await gptscriptClient.writeFileInWorkspace(
    ".metadata.json",
    Buffer.from(JSON.stringify(output, null, 2))
  );
}

main()
  .then(() => process.exit(0))
  .catch((err) => {
    console.log(JSON.stringify({ error: err.message }));
    process.exit(0);
  });
