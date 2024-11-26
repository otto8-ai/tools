import { Client } from "@notionhq/client";
import dotenv from "dotenv";
import path from "path";
import {
    PageObjectResponse,
    SearchResponse,
    ListBlockChildrenResponse,
} from "@notionhq/client/build/src/api-endpoints";
import { getPageContent } from "./page";

dotenv.config();

interface OutputMetadata {
    files: {
        [pageUrl: string]: {
            updatedAt: string;
            filePath: string;
            url: string;
            sizeInBytes: number;
        };
    };
    status: string;
}

async function writePageToFile(
    path: string,
    content: string,
    gptscriptClient: any
): Promise<number> {
    const buffer = Buffer.from(content);
    await gptscriptClient.writeFileInWorkspace(path, buffer);
    return buffer.length;
}

function getPath(page: PageObjectResponse, folderPath: string): string {
    const pageId = page.id;
    const fileDir = path.join(folderPath, pageId.toString());
    let title = getTitle(page);
    return path.join(fileDir, title + ".md");
}

function getTitle(page: any): string {
    if (page.type === "child_database") {
        return page.child_database.title;
    }

    if (page.object !== "page") {
        return "";
    }
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

async function getAllPagesIteratively(
    client: Client,
    output: OutputMetadata,
    gptscriptClient: any
) {
    const pages = new Map<
        string,
        {
            page: PageObjectResponse;
            path: string;
            hasChildren: boolean;
        }
    >();
    const stack: (PageObjectResponse | null)[] = [];

    let cursor = null;
    do {
        const response: SearchResponse = await client.search({
            page_size: 100,
            start_cursor: cursor ?? undefined,
            filter: { property: "object", value: "page" },
        });

        stack.push(...(response.results as PageObjectResponse[]));

        cursor = response.has_more ? response.next_cursor : null;
    } while (cursor);

    while (stack.length > 0) {
        const currentPage = stack.pop();
        if (!currentPage) {
            continue;
        }

        if (pages.has(currentPage.id)) {
            continue;
        }
        if (currentPage.url) {
            output.status = `Syncing page ${currentPage.url}...`;
            await gptscriptClient.writeFileInWorkspace(
                ".metadata.json",
                Buffer.from(JSON.stringify(output, null, 2))
            );
        }
        pages.set(currentPage.id, {
            page: currentPage,
            path: "",
            hasChildren: false,
        });
        let childCursor = null;
        do {
            try {
                const childResponse: ListBlockChildrenResponse =
                    await client.blocks.children.list({
                        block_id: currentPage.id,
                        page_size: 100,
                        start_cursor: childCursor ?? undefined,
                    });

                const hasChildren = childResponse.results.some(
                    (child) => (child as any).type === "child_page"
                );
                pages.set(currentPage.id, {
                    page: currentPage,
                    path: "",
                    hasChildren: hasChildren,
                });

                for (const child of childResponse.results) {
                    if (!pages.has(child.id)) {
                        try {
                            if ((child as any).type === "child_page") {
                                const childPage = await getPage(
                                    client,
                                    child.id
                                );
                                stack.push(childPage);
                            } else if ((child as any).has_children) {
                                stack.push(child as any);
                            } else if (
                                (child as any).type === "child_database"
                            ) {
                                stack.push(child as any);
                            }
                        } catch (err: any) {
                            console.error(
                                `Failed to get page ${child.id}: ${err.message}`
                            );
                        }
                    }
                }

                childCursor = childResponse.has_more
                    ? childResponse.next_cursor
                    : null;
            } catch (err: any) {
                console.error(
                    `Failed to get children for block ${currentPage.id}: ${err.message}`
                );
                break;
            }
        } while (childCursor);
    }

    for (const pageData of pages.values()) {
        let folderPath = "";
        if (pageData.hasChildren) {
            folderPath = getTitle(pageData.page);
        }
        let p = pageData.page;
        while (p.parent) {
            try {
                let parentId = "";
                if (p.parent.type === "page_id") {
                    parentId = p.parent.page_id;
                } else if (p.parent.type === "block_id") {
                    parentId = p.parent.block_id;
                } else if (p.parent.type === "database_id") {
                    parentId = p.parent.database_id;
                }
                const parentPage = pages.get(parentId)?.page;
                if (!parentPage) {
                    break;
                }
                const parentTitle = getTitle(parentPage);
                if (parentTitle) {
                    folderPath = path.join(parentTitle, folderPath);
                }
                p = parentPage;
            } catch (err: any) {
                folderPath = "";
                break;
            }
        }
        pageData.path = getPath(pageData.page, folderPath);
    }

    for (const [pageId, pageData] of pages.entries()) {
        if (pageData.page.object !== "page") {
            pages.delete(pageId);
        }
    }
    return pages;
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
        metadataFile = await gptscriptClient.readFileInWorkspace(
            ".metadata.json"
        );
    } catch (err: any) {
        // Ignore any error if the metadata file doesn't exist. Ideally we should check for only not existing error but sdk doesn't provide that
    }
    if (metadataFile) {
        output = JSON.parse(metadataFile.toString());
    }

    if (!output.files) {
        output.files = {};
    }

    let syncedCount = 0;
    const allPages = await getAllPagesIteratively(
        client,
        output,
        gptscriptClient
    );

    for (const [pageId, { page, path }] of allPages.entries()) {
        if (
            !output.files[pageId] ||
            output.files[pageId].updatedAt !== page.last_edited_time
        ) {
            console.error(`Writing page url: ${page.url}`);
            const content = await getPageContent(client, pageId);
            const sizeInBytes = await writePageToFile(
                path,
                content,
                gptscriptClient
            );
            output.files[pageId] = {
                url: page.url,
                filePath: path,
                updatedAt: page.last_edited_time,
                sizeInBytes: sizeInBytes,
            };
            syncedCount++;
        }
        output.status = `${syncedCount}/${allPages.size} number of pages have been synced`;
        await gptscriptClient.writeFileInWorkspace(
            ".metadata.json",
            Buffer.from(JSON.stringify(output, null, 2))
        );
    }
    for (const [pageId, fileInfo] of Object.entries(output.files)) {
        if (!allPages.has(pageId)) {
            try {
                await gptscriptClient.deleteFileInWorkspace(fileInfo.filePath);
                delete output.files[pageId];
                console.error(`Deleted file and entry for page ID: ${pageId}`);
            } catch (error) {
                console.error(
                    `Failed to delete file ${fileInfo.filePath}:`,
                    error
                );
            }
        }
    }

    output.status = "";
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
