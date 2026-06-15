import { FormEvent, useEffect, useRef, useState } from "react";
import {
  Link,
  LinkStats,
  MSG_PROFANITY,
  MSG_INVALID_URL,
  createLink,
  deleteLink,
  getClickCount,
  getStats,
  hasProfanity,
  isValidUrl,
  listLinks,
  login,
  register,
  resolveShortUrl
} from "./api";
import "./styles.css";

const CREATE_TIMEOUT_MS = 3_000;
const ERROR_MS = 3_000;

function displayShortUrl(shortUrl: string): string {
  try {
    const { hostname, pathname } = new URL(shortUrl);
    const code = pathname.replace(/^\//, "");
    const host =
      hostname.length > 22
        ? `${hostname.slice(0, 10)}…${hostname.slice(-10)}`
        : hostname;
    return `${host}/${code}`;
  } catch {
    return shortUrl;
  }
}

function App() {
  const [token, setToken] = useState(() => localStorage.getItem("token") || "");
  const [email, setEmail] = useState(
    () => localStorage.getItem("userEmail") || ""
  );
  const [password, setPassword] = useState("");
  const [url, setUrl] = useState("");
  const [links, setLinks] = useState<Link[]>([]);
  const [stats, setStats] = useState<LinkStats | null>(null);
  const [statsLinkId, setStatsLinkId] = useState<string | null>(null);
  const [error, setError] = useState("");
  const [creating, setCreating] = useState(false);
  const errorTimerRef = useRef<number | null>(null);

  const isAuthed = token !== "";

  function showError(message: string) {
    if (errorTimerRef.current !== null) {
      window.clearTimeout(errorTimerRef.current);
    }
    setError(message);
    errorTimerRef.current = window.setTimeout(() => {
      setError("");
      errorTimerRef.current = null;
    }, ERROR_MS);
  }

  useEffect(() => {
    return () => {
      if (errorTimerRef.current !== null) {
        window.clearTimeout(errorTimerRef.current);
      }
    };
  }, []);

  async function run(action: () => Promise<void>) {
    try {
      await action();
    } catch (err) {
      showError(err instanceof Error ? err.message : "Unknown error");
    }
  }

  async function refreshLinks() {
    if (!token) return;
    const items = await listLinks(token);
    setLinks(items);
  }

  useEffect(() => {
    void refreshLinks();
  }, [token]);

  async function onRegister(event: FormEvent) {
    event.preventDefault();
    await run(async () => {
      const result = await register(email, password);
      localStorage.setItem("token", result.token);
      localStorage.setItem("userEmail", email);
      setToken(result.token);
    });
  }

  async function onLogin(event: FormEvent) {
    event.preventDefault();
    await run(async () => {
      const result = await login(email, password);
      localStorage.setItem("token", result.token);
      localStorage.setItem("userEmail", email);
      setToken(result.token);
    });
  }

  async function onCreate(event: FormEvent) {
    event.preventDefault();
    if (creating) return;

    if (hasProfanity(url)) {
      showError(MSG_PROFANITY);
      return;
    }

    if (!isValidUrl(url)) {
      showError(MSG_INVALID_URL);
      return;
    }

    setCreating(true);
    const timeoutId = window.setTimeout(() => setCreating(false), CREATE_TIMEOUT_MS);

    try {
      await createLink(token, url);
      setUrl("");
      await refreshLinks();
    } catch (err) {
      showError(err instanceof Error ? err.message : "Unknown error");
    } finally {
      window.clearTimeout(timeoutId);
      setCreating(false);
    }
  }

  function logout() {
    localStorage.removeItem("token");
    localStorage.removeItem("userEmail");
    setToken("");
    setLinks([]);
    setStats(null);
    setStatsLinkId(null);
  }

  async function copyShortUrl(shortUrl: string) {
    await navigator.clipboard.writeText(shortUrl);
  }

  return (
    <main className="page">
      <section className="card">
        <header className="app-header">
          <div className="app-header-main">
            <h1>Shortify</h1>
            <p className="hint">
              {isAuthed
                ? "Paste a link and get a short version."
                : "Sign in to shorten links and view click stats."}
            </p>
          </div>
          {isAuthed && (
            <div className="app-header-user">
              <span className="user-label">{email}</span>
              <button onClick={logout}>Logout</button>
            </div>
          )}
        </header>

        {error && <p className="error">{error}</p>}

        {isAuthed && <hr className="header-divider" />}

        {!isAuthed ? (
          <form className="form">
            <input
              value={email}
              onChange={(event) => setEmail(event.target.value)}
              placeholder="email"
              type="email"
              autoComplete="email"
            />
            <p className="field-hint">Use a valid email, e.g. name@example.com</p>
            <input
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              placeholder="password"
              type="password"
              minLength={6}
            />
            <p className="field-hint">Password must be at least 6 characters.</p>
            <div className="row">
              <button onClick={onRegister}>Register</button>
              <button onClick={onLogin} type="button">
                Login
              </button>
            </div>
          </form>
        ) : (
          <>
            <form className="form form-create" onSubmit={onCreate}>
              <input
                value={url}
                onChange={(event) => setUrl(event.target.value)}
                placeholder="https://example.com"
              />
              <button disabled={creating} type="submit">
                Create short link
              </button>
            </form>

            <div className="table-wrap">
            <table className="links-table">
              <colgroup>
                <col className="col-short" />
                <col className="col-original" />
                <col className="col-clicks" />
                <col className="col-actions" />
              </colgroup>
              <thead>
                <tr>
                  <th>Short</th>
                  <th>Original</th>
                  <th className="clicks-cell">Clicks</th>
                  <th className="actions-header">Actions</th>
                </tr>
              </thead>
              <tbody>
                {links.map((link) => {
                  const shortUrl = resolveShortUrl(link);

                  return (
                  <tr key={link.id}>
                    <td className="url-cell">
                      <div className="short-cell">
                        <a
                          href={shortUrl}
                          className="short-link"
                          rel="noreferrer"
                          target="_blank"
                          title={shortUrl}
                        >
                          {displayShortUrl(shortUrl)}
                        </a>
                        <button
                          type="button"
                          className="btn-secondary"
                          onClick={() => run(async () => copyShortUrl(shortUrl))}
                        >
                          Copy
                        </button>
                      </div>
                    </td>
                    <td className="url-cell" title={link.original_url}>
                      <span className="url-text">{link.original_url}</span>
                    </td>
                    <td className="clicks-cell">{getClickCount(link)}</td>
                    <td className="actions-cell">
                      <div className="actions">
                      <button
                        onClick={() =>
                          run(async () => {
                            if (statsLinkId === link.id) {
                              setStats(null);
                              setStatsLinkId(null);
                              return;
                            }
                            setStats(await getStats(token, link.id));
                            setStatsLinkId(link.id);
                          })
                        }
                      >
                        Stats
                      </button>
                      <button
                        onClick={() =>
                          run(async () => {
                            await deleteLink(token, link.id);
                            if (statsLinkId === link.id) {
                              setStats(null);
                              setStatsLinkId(null);
                            }
                            await refreshLinks();
                          })
                        }
                      >
                        Delete
                      </button>
                      </div>
                    </td>
                  </tr>
                  );
                })}
              </tbody>
            </table>
            </div>

            {stats && (
              <section className="stats">
                <h2 className="stats-title">Click stats</h2>
                <div className="stats-total">
                  <span className="stats-total-value">{stats.total_clicks}</span>
                  <span className="stats-total-label">total clicks</span>
                </div>
                {stats.recent.length > 0 ? (
                  <>
                    <h3 className="stats-subtitle">Recent activity</h3>
                    <ul className="stats-list">
                      {stats.recent.map((click, index) => {
                        const date = new Date(click.created_at);
                        return (
                          <li key={index} className="stats-item">
                            <span className="stats-date">{date.toLocaleDateString()}</span>
                            <span className="stats-time">{date.toLocaleTimeString()}</span>
                          </li>
                        );
                      })}
                    </ul>
                  </>
                ) : (
                  <p className="stats-empty">No clicks yet.</p>
                )}
              </section>
            )}
          </>
        )}
      </section>
    </main>
  );
}

export default App;
