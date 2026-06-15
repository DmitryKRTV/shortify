import { FormEvent, useEffect, useState } from "react";
import {
  Link,
  LinkStats,
  createLink,
  deleteLink,
  getClickCount,
  getStats,
  listLinks,
  login,
  register,
  resolveShortUrl
} from "./api";
import "./styles.css";

function App() {
  const [token, setToken] = useState(() => localStorage.getItem("token") || "");
  const [email, setEmail] = useState("demo@example.com");
  const [password, setPassword] = useState("secret123");
  const [url, setUrl] = useState("");
  const [links, setLinks] = useState<Link[]>([]);
  const [stats, setStats] = useState<LinkStats | null>(null);
  const [statsLinkId, setStatsLinkId] = useState<string | null>(null);
  const [error, setError] = useState("");

  const isAuthed = token !== "";

  async function run(action: () => Promise<void>) {
    setError("");
    try {
      await action();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unknown error");
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
      setToken(result.token);
    });
  }

  async function onLogin(event: FormEvent) {
    event.preventDefault();
    await run(async () => {
      const result = await login(email, password);
      localStorage.setItem("token", result.token);
      setToken(result.token);
    });
  }

  async function onCreate(event: FormEvent) {
    event.preventDefault();
    await run(async () => {
      await createLink(token, url);
      setUrl("");
      await refreshLinks();
    });
  }

  function logout() {
    localStorage.removeItem("token");
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
        <h1>Shortify</h1>
        <p className="hint">
          {isAuthed
            ? "Paste a link and get a short version."
            : "Sign in to shorten links and view click stats."}
        </p>

        {error && <pre className="error">{error}</pre>}

        {!isAuthed ? (
          <form className="form">
            <input
              value={email}
              onChange={(event) => setEmail(event.target.value)}
              placeholder="email"
            />
            <input
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              placeholder="password"
              type="password"
            />
            <div className="row">
              <button onClick={onRegister}>Register</button>
              <button onClick={onLogin} type="button">
                Login
              </button>
            </div>
          </form>
        ) : (
          <>
            <div className="row between">
              <span>Authorized</span>
              <button onClick={logout}>Logout</button>
            </div>

            <form className="form" onSubmit={onCreate}>
              <input
                value={url}
                onChange={(event) => setUrl(event.target.value)}
                placeholder="https://example.com"
              />
              <button>Create short link</button>
            </form>

            <table>
              <thead>
                <tr>
                  <th>Short</th>
                  <th>Original</th>
                  <th>Clicks</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {links.map((link) => {
                  const shortUrl = resolveShortUrl(link);

                  return (
                  <tr key={link.id}>
                    <td>
                      <div className="short-cell">
                        <a
                          href={shortUrl}
                          className="short-link"
                          rel="noreferrer"
                          target="_blank"
                        >
                          {shortUrl}
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
                    <td>{link.original_url}</td>
                    <td>{getClickCount(link)}</td>
                    <td className="actions">
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
                    </td>
                  </tr>
                  );
                })}
              </tbody>
            </table>

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
