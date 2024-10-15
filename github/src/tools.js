import {GPTScript} from "@gptscript-ai/gptscript";

export async function searchIssuesAndPRs(octokit, owner, repo, query, perPage = 100, page = 1) {
    let q = '';

    if (owner) {
        const { data: { type } } = await octokit.users.getByUsername({ username: owner });
        const ownerQualifier = type === 'User' ? `user:${owner}` : `org:${owner}`;
        q = repo ? `repo:${owner}/${repo}` : ownerQualifier;
    } else if (repo) {
        throw new Error('Repository given without an owner. Please provide an owner.');
    } else {
        throw new Error('Owner and repository must be provided.');
    }

    if (query) {
        q += ` ${query}`;
    }

    const { data: { items } } = await octokit.search.issuesAndPullRequests({
        q: q.trim(),
        per_page: perPage,
        page: page
    });

    if (items.length > 10) {
        const gptscriptClient = new GPTScript();
        const dataset = await gptscriptClient.createDataset(process.env.GPTSCRIPT_WORKSPACE_DIR, `${query}_github_issues_prs`, `Search results for ${query} on GitHub`);
        for (const issue of items) {
            await gptscriptClient.addDatasetElement(
                process.env.GPTSCRIPT_WORKSPACE_DIR,
                dataset.id,
                `${issue.id}`,
                '',
                `#${issue.number} - ${issue.title} (ID: ${issue.id}) - ${issue.html_url}`
            );
        }
        console.log(`Created dataset with ID ${dataset.id} with ${items.length} results`);
        return;
    }

    items.forEach(issue => {
        console.log(`#${issue.number} - ${issue.title} (ID: ${issue.id}) - ${issue.html_url}`);
    });
}

export async function getIssue(octokit, owner, repo, issueNumber) {
    const { data } = await octokit.issues.get({
        owner,
        repo,
        issue_number: issueNumber,
    });
    console.log(data);
    console.log(`https://github.com/${owner}/${repo}/issues/${issueNumber}`);
}

export async function createIssue(octokit, owner, repo, title, body) {
    const issue = await octokit.issues.create({
        owner,
        repo,
        title,
        body
    });

    console.log(`Created issue #${issue.data.number} - ${issue.data.title} (ID: ${issue.data.id}) - https://github.com/${owner}/${repo}/issues/${issue.data.number}`);
}

export async function modifyIssue(octokit, owner, repo, issueNumber, title, body) {
    const issue = await octokit.issues.update({
        owner,
        repo,
        issue_number: issueNumber,
        title,
        body
    });

    console.log(`Modified issue #${issue.data.number} - ${issue.data.title} (ID: ${issue.data.id}) - https://github.com/${owner}/${repo}/issues/${issue.data.number}`);
}

export async function closeIssue(octokit, owner, repo, issueNumber) {
    await octokit.issues.update({
        owner,
        repo,
        issue_number: issueNumber,
        state: 'closed'
    });
    console.log(`Closed issue #${issueNumber} - https://github.com/${owner}/${repo}/issues/${issueNumber}`);
}

export async function listIssueComments(octokit, owner, repo, issueNumber) {
    const { data } = await octokit.issues.listComments({
        owner,
        repo,
        issue_number: issueNumber,
    });

    if (data.length > 10) {
        const gptscriptClient = new GPTScript();
        const dataset = await gptscriptClient.createDataset(
            process.env.GPTSCRIPT_WORKSPACE_DIR,
            `${owner}_${repo}_issue_${issueNumber}_comments`,
            `Comments for issue #${issueNumber} in ${owner}/${repo}`
        );
        for (const comment of data) {
            await gptscriptClient.addDatasetElement(
                process.env.GPTSCRIPT_WORKSPACE_DIR,
                dataset.id,
                `${comment.id}`,
                '',
                `Comment by ${comment.user.login}: ${comment.body} - https://github.com/${owner}/${repo}/issues/${issueNumber}#issuecomment-${comment.id}`
            );
        }
        console.log(`Created dataset with ID ${dataset.id} with ${data.length} comments`);
        return;
    }

    data.forEach(comment => {
        console.log(`Comment by ${comment.user.login}: ${comment.body} - https://github.com/${owner}/${repo}/issues/${issueNumber}#issuecomment-${comment.id}`);
    });
}

export async function addCommentToIssue(octokit, owner, repo, issueNumber, comment) {
    const issueComment = await octokit.issues.createComment({
        owner,
        repo,
        issue_number: issueNumber,
        body: comment
    });

    console.log(`Added comment to issue #${issueNumber}: ${issueComment.data.body} - https://github.com/${owner}/${repo}/issues/${issueNumber}`);
}

export async function getPR(octokit, owner, repo, prNumber) {
    const { data } = await octokit.pulls.get({
        owner,
        repo,
        pull_number: prNumber,
    });
    console.log(data);
    console.log(`https://github.com/${owner}/${repo}/pull/${prNumber}`);
}

export async function createPR(octokit, owner, repo, title, body, head, base) {
    const pr = await octokit.pulls.create({
        owner,
        repo,
        title,
        body,
        head,
        base
    });

    console.log(`Created PR #${pr.data.number} - ${pr.data.title} (ID: ${pr.data.id}) - https://github.com/${owner}/${repo}/pull/${pr.data.number}`);
}

export async function modifyPR(octokit, owner, repo, prNumber, title, body) {
    const pr = await octokit.pulls.update({
        owner,
        repo,
        pull_number: prNumber,
        title,
        body
    });

    console.log(`Modified PR #${pr.data.number} - ${pr.data.title} (ID: ${pr.data.id}) - https://github.com/${owner}/${repo}/pull/${pr.data.number}`);
}

export async function closePR(octokit, owner, repo, prNumber) {
    await octokit.pulls.update({
        owner,
        repo,
        pull_number: prNumber,
        state: 'closed'
    });

    console.log(`Deleted PR #${prNumber} - https://github.com/${owner}/${repo}/pull/${prNumber}`);
}

export async function listPRComments(octokit, owner, repo, prNumber) {
    const { data } = await octokit.issues.listComments({
        owner,
        repo,
        issue_number: prNumber,
    });

    if (data.length > 10) {
        const gptscriptClient = new GPTScript();
        const dataset = await gptscriptClient.createDataset(
            process.env.GPTSCRIPT_WORKSPACE_DIR,
            `${owner}_${repo}_pr_${prNumber}_comments`,
            `Comments for PR #${prNumber} in ${owner}/${repo}`
        );
        for (const comment of data) {
            await gptscriptClient.addDatasetElement(
                process.env.GPTSCRIPT_WORKSPACE_DIR,
                dataset.id,
                `${comment.id}`,
                '',
                `Comment by ${comment.user.login}: ${comment.body} - https://github.com/${owner}/${repo}/pull/${prNumber}#issuecomment-${comment.id}`
            );
        }
        console.log(`Created dataset with ID ${dataset.id} with ${data.length} comments`);
        return;
    }

    data.forEach(comment => {
        console.log(`Comment by ${comment.user.login}: ${comment.body} - https://github.com/${owner}/${repo}/pull/${prNumber}#issuecomment-${comment.id}`);
    });
}

export async function addCommentToPR(octokit, owner, repo, prNumber, comment) {
    const prComment = await octokit.issues.createComment({
        owner,
        repo,
        issue_number: prNumber,
        body: comment
    });

    console.log(`Added comment to PR #${prNumber}: ${prComment.data.body} - https://github.com/${owner}/${repo}/pull/${prNumber}`);
}


export async function listRepos(octokit, owner) {
    const repos = await octokit.repos.listForUser({
        username: owner,
        per_page: 100
    });

    if (repos.data.length > 10) {
        const gptscriptClient = new GPTScript();
        const dataset = await gptscriptClient.createDataset(
            process.env.GPTSCRIPT_WORKSPACE_DIR,
            `${owner}_github_repos`,
            `GitHub repos for ${owner}`
        );
        for (const repo of repos.data) {
            await gptscriptClient.addDatasetElement(
                process.env.GPTSCRIPT_WORKSPACE_DIR,
                dataset.id,
                `${repo.id}`,
                '',
                `${repo.name} (ID: ${repo.id}) - https://github.com/${owner}/${repo.name}`
            );
        }
        console.log(`Created dataset with ID ${dataset.id} with ${repos.data.length} repositories`);
        return;
    }

    repos.data.forEach(repo => {
        console.log(`${repo.name} (ID: ${repo.id}) - https://github.com/${owner}/${repo.name}`);
    });
}

export async function getStarCount(octokit, owner, repo) {
    const { data } = await octokit.repos.get({
        owner,
        repo,
    });
    console.log(data.stargazers_count);
}
