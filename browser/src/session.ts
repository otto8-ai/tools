import fs from 'node:fs/promises'
import os from 'node:os'
import path from 'node:path'
import { type IncomingHttpHeaders } from 'node:http'
import { createHash } from 'node:crypto'
import { type BrowserContext, type Page } from 'playwright'
import { newBrowserContext } from './context.ts'
import { Mutex } from 'async-mutex'

const APP_CACHE_DIR = (() => {
  const homeDir = os.homedir()
  const appPath = path.join('obot', 'tools', 'browser')

  switch (os.platform()) {
    case 'win32':
      return path.join(process.env.APPDATA ?? path.join(homeDir, 'AppData', 'Roaming'), appPath)
    case 'darwin':
      return path.join(homeDir, 'Library', 'Caches', appPath)
    default:
      return path.join(process.env.XDG_CACHE_HOME ?? path.join(homeDir, '.cache'), appPath)
  }
})()

async function clearAppCacheDir(): Promise<void> {
  try {
    await fs.rm(APP_CACHE_DIR, { recursive: true, force: true })
    console.log(`Cleared APP_CACHE_DIR at startup: ${APP_CACHE_DIR}`)
  } catch (error) {
    console.error(`Failed to clear APP_CACHE_DIR: ${error}`)
  }
}

// Call the function at startup
await clearAppCacheDir()

let sessionManager: SessionManager | undefined

interface ManagedSession {
  session: Session
  cleanupTimeout?: NodeJS.Timeout
}

export class SessionManager {
  private readonly sessions = new Map<string, ManagedSession>()
  private readonly sessionsLock: Mutex = new Mutex()

  private constructor () {
  }

  static async create (): Promise<SessionManager> {
    sessionManager ??= new SessionManager()
    return sessionManager
  }

  async withSession (sessionId: string, fn: (browserContext: BrowserContext, openPages: Map<string, Page>) => Promise<void>): Promise<void> {
    let managedSession: ManagedSession | undefined
    await this.sessionsLock.runExclusive(async () => {
      managedSession = this.sessions.get(sessionId)
      if (!managedSession) {
        managedSession = { session: await Session.create(sessionId) }
        this.sessions.set(sessionId, managedSession)
      }
      if (managedSession.cleanupTimeout != null) clearTimeout(managedSession.cleanupTimeout)
    })

    await managedSession?.session.lock.runExclusive(async () => {
      if (managedSession?.session.browserContext != null) {
        await fn(managedSession.session.browserContext, managedSession.session.openPages)
        managedSession.cleanupTimeout = setTimeout(() => {
          void this.deleteSession(sessionId)
        }, SESSION_TTL)
      }
    })
  }

  private async deleteSession (sessionId: string): Promise<void> {
    await this.sessionsLock.runExclusive(async () => {
      const managedSession = this.sessions.get(sessionId)
      if (managedSession) {
        const { session, cleanupTimeout } = managedSession
        if (cleanupTimeout != null) clearTimeout(cleanupTimeout)
        await session?.close()
        this.sessions.delete(sessionId)
      }
    })
  }
}

const SESSION_TTL = 5 * 60 * 1000 // 5 minutes

class Session {
  sessionId: string
  sessionDir: string = ''
  browserContext?: BrowserContext
  openPages = new Map<string, Page>()
  lock: Mutex = new Mutex()

  private constructor (sessionId: string) {
    this.sessionId = sessionId
  }

  static async create (sessionId: string): Promise<Session> {
    const session = new Session(sessionId)
    session.sessionDir = await mkSessionDir(sessionId)
    session.browserContext = await newBrowserContext(session.sessionDir)
    return session
  }

  async close (): Promise<void> {
    await this.browserContext?.close()
    await fs.rm(this.sessionDir, { recursive: true })
  }
}

async function mkSessionDir (sessionId: string): Promise<string> {
  const sessionDir = path.resolve(APP_CACHE_DIR, 'browser_sessions', sessionId)
  await fs.mkdir(sessionDir, { recursive: true })
  return sessionDir
}

export function getSessionId (headers: IncomingHttpHeaders): string {
  const workspaceId = getGPTScriptEnv(headers, 'GPTSCRIPT_WORKSPACE_ID')
  if (workspaceId == null) throw new Error('No GPTScript workspace ID provided')

  return createHash('sha256').update(workspaceId).digest('hex').substring(0, 16)
}

export function getWorkspaceId (headers: IncomingHttpHeaders): string | undefined {
  return getGPTScriptEnv(headers, 'GPTSCRIPT_WORKSPACE_ID')
}

export function getGPTScriptEnv (headers: IncomingHttpHeaders, envKey: string): string | undefined {
  const envHeader = headers?.['x-gptscript-env']
  const envArray = Array.isArray(envHeader) ? envHeader : [envHeader]

  for (const env of envArray) {
    if (env == null) {
      continue
    }

    for (const pair of env.split(',')) {
      const [key, value] = pair.split('=').map(part => part.trim())
      if (key === envKey) {
        return value
      }
    }
  }

  return undefined
}
