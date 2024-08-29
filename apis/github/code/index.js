import { Octokit } from '@octokit/rest';
import { listIssues, listPRs, createIssue, modifyIssue, deleteIssue, searchIssues, createPR, modifyPR, deletePR, addCommentToIssue, addCommentToPR, getIssue, getPR, getIssueComments, getPRComments, listRepos, closeIssue } from './src/tools.js';

if (process.argv.length !== 3) {
    console.log('Usage: node index.js <command>');
    process.exit(1);
}

const command = process.argv[2];
const token = process.env.GITHUB_TOKEN;

if (!token) {
    console.log('GITHUB_TOKEN environment variable must be set.');
    process.exit(1);
}

const octokit = new Octokit({ auth: token });

try {
    switch (command) {
        case 'listIssues':
            await listIssues(octokit, process.env.OWNER, process.env.REPO);
            break;
        case 'listPRs':
            await listPRs(octokit, process.env.OWNER, process.env.REPO);
            break;
        case 'createIssue':
            await createIssue(octokit, process.env.OWNER, process.env.REPO, process.env.TITLE, process.env.BODY);
            break;
        case 'modifyIssue':
            await modifyIssue(octokit, process.env.OWNER, process.env.REPO, process.env.ISSUENUMBER, process.env.NEWTITLE, process.env.NEWBODY);
            break;
        case 'deleteIssue':
            await deleteIssue(octokit, process.env.OWNER, process.env.REPO, process.env.ISSUENUMBER);
            break;
        case 'searchIssues':
            await searchIssues(octokit, process.env.OWNER, process.env.REPO, process.env.QUERY);
            break;
        case 'createPR':
            await createPR(octokit, process.env.OWNER, process.env.REPO, process.env.TITLE, process.env.BODY, process.env.HEAD, process.env.BASE);
            break;
        case 'modifyPR':
            await modifyPR(octokit, process.env.OWNER, process.env.REPO, process.env.PRNUMBER, process.env.NEWTITLE, process.env.NEWBODY);
            break;
        case 'deletePR':
            await deletePR(octokit, process.env.OWNER, process.env.REPO, process.env.PRNUMBER);
            break;
        case 'addCommentToIssue':
            await addCommentToIssue(octokit, process.env.OWNER, process.env.REPO, process.env.ISSUENUMBER, process.env.COMMENT);
            break;
        case 'addCommentToPR':
            await addCommentToPR(octokit, process.env.OWNER, process.env.REPO, process.env.PRNUMBER, process.env.COMMENT);
            break;
        case 'getIssue':
            await getIssue(octokit, process.env.OWNER, process.env.REPO, process.env.ISSUENUMBER);
            break;
        case 'getPR':
            await getPR(octokit, process.env.OWNER, process.env.REPO, process.env.PRNUMBER);
            break;
        case 'getIssueComments':
            await getIssueComments(octokit, process.env.OWNER, process.env.REPO, process.env.ISSUENUMBER);
            break;
        case 'getPRComments':
            await getPRComments(octokit, process.env.OWNER, process.env.REPO, process.env.PRNUMBER);
            break;
        case 'listRepos':
            await listRepos(octokit, process.env.OWNER);
            break;
        case 'closeIssue':
            await closeIssue(octokit, process.env.OWNER, process.env.REPO, process.env.ISSUENUMBER);
            break;
        default:
            throw new Error(`Unknown command: ${command}`);
    }
} catch (error) {
    console.log(`Error running ${command}: ${error.message}`);
    process.exit(1);
}
