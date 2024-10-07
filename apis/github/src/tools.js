export async function closeIssue(octokit, owner, repo, issueNumber) {
    await octokit.issues.update({
        owner,
        repo,
        issue_number: issueNumber,
        state: 'closed'
    });
    console.log(`Closed issue #${issueNumber} - https://github.com/${owner}/${repo}/issues/${issueNumber}`);
}

export async function listRepos(octokit, owner) {
    const repos = await octokit.repos.listForUser({
        username: owner,
        per_page: 100
    });

    repos.data.forEach(repo => {
        console.log(`${repo.name} (ID: ${repo.id}) - https://github.com/${owner}/${repo.name}`);
    });
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

export async function deleteIssue(octokit, owner, repo, issueNumber) {
    await octokit.issues.update({
        owner,
        repo,
        issue_number: issueNumber,
        state: 'closed'
    });

    console.log(`Deleted issue #${issueNumber} - https://github.com/${owner}/${repo}/issues/${issueNumber}`);
}

export async function searchIssuesAndPRs(octokit, owner, repo, query) {
    let q = '';

    if (owner) {
        // Determine if the owner is a user or an organization
        const { data: { type } } = await octokit.users.getByUsername({ username: owner });

        // Construct the query based on the presence of repo and owner type
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

    // Search for issues and pull requests
    const { data: { items } } = await octokit.search.issuesAndPullRequests({ q: q.trim() });

    // Log the results
    items.forEach(issue => {
        console.log(`#${issue.number} - ${issue.title} (ID: ${issue.id}) - ${issue.html_url}`);
    });
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

export async function deletePR(octokit, owner, repo, prNumber) {
    await octokit.pulls.update({
        owner,
        repo,
        pull_number: prNumber,
        state: 'closed'
    });

    console.log(`Deleted PR #${prNumber} - https://github.com/${owner}/${repo}/pull/${prNumber}`);
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

export async function addCommentToPR(octokit, owner, repo, prNumber, comment) {
    const prComment = await octokit.issues.createComment({
        owner,
        repo,
        issue_number: prNumber,
        body: comment
    });

    console.log(`Added comment to PR #${prNumber}: ${prComment.data.body} - https://github.com/${owner}/${repo}/pull/${prNumber}`);
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

export async function getPR(octokit, owner, repo, prNumber) {
    const { data } = await octokit.pulls.get({
        owner,
        repo,
        pull_number: prNumber,
    });
    console.log(data);
    console.log(`https://github.com/${owner}/${repo}/pull/${prNumber}`);
}

export async function getIssueComments(octokit, owner, repo, issueNumber) {
    const { data } = await octokit.issues.listComments({
        owner,
        repo,
        issue_number: issueNumber,
    });
    data.forEach(comment => {
        console.log(`Comment by ${comment.user.login}: ${comment.body} - https://github.com/${owner}/${repo}/issues/${issueNumber}#issuecomment-${comment.id}`);
    });
}

export async function getPRComments(octokit, owner, repo, prNumber) {
    const { data } = await octokit.issues.listComments({
        owner,
        repo,
        issue_number: prNumber,
    });
    data.forEach(comment => {
        console.log(`Comment by ${comment.user.login}: ${comment.body} - https://github.com/${owner}/${repo}/pull/${prNumber}#issuecomment-${comment.id}`);
    });
}

export async function getStarCount(octokit, owner, repo) {
    const { data } = await octokit.repos.get({
        owner,
        repo,
    });
    console.log(data.stargazers_count);
}