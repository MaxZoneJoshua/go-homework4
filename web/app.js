const statusEl = document.getElementById("status");
const baseUrlInput = document.getElementById("baseUrl");
const registerForm = document.getElementById("registerForm");
const loginForm = document.getElementById("loginForm");
const logoutBtn = document.getElementById("logoutBtn");
const sessionEl = document.getElementById("session");
const sessionLabel = document.getElementById("sessionLabel");
const postForm = document.getElementById("postForm");
const savePostBtn = document.getElementById("savePostBtn");
const cancelEditBtn = document.getElementById("cancelEditBtn");
const postsEl = document.getElementById("posts");
const refreshPostsBtn = document.getElementById("refreshPosts");
const postDetailsEl = document.getElementById("postDetails");
const commentSectionEl = document.getElementById("commentSection");
const commentsEl = document.getElementById("comments");
const commentForm = document.getElementById("commentForm");

const state = {
  token: null,
  user: null,
  editingPostId: null,
  selectedPostId: null,
  posts: []
};

function setStatus(message, tone = "info") {
  statusEl.textContent = message;
  statusEl.classList.remove("error", "success");
  if (tone === "error") {
    statusEl.classList.add("error");
  }
  if (tone === "success") {
    statusEl.classList.add("success");
  }
}

function getBaseUrl() {
  const raw = baseUrlInput.value.trim();
  const stored = localStorage.getItem("blogBaseUrl");
  if (!raw && stored) {
    baseUrlInput.value = stored;
    return stored;
  }
  if (!raw) {
    const origin = window.location.origin;
    if (origin && origin !== "null") {
      return origin;
    }
    return "";
  }
  const cleaned = raw.replace(/\/$/, "");
  localStorage.setItem("blogBaseUrl", cleaned);
  return cleaned;
}

function saveAuth() {
  const payload = { token: state.token, user: state.user };
  localStorage.setItem("blogAuth", JSON.stringify(payload));
}

function loadAuth() {
  const raw = localStorage.getItem("blogAuth");
  if (!raw) {
    return;
  }
  try {
    const payload = JSON.parse(raw);
    state.token = payload.token;
    state.user = payload.user;
  } catch (err) {
    localStorage.removeItem("blogAuth");
  }
}

function clearAuth() {
  state.token = null;
  state.user = null;
  saveAuth();
}

async function safeJson(response) {
  const text = await response.text();
  if (!text) {
    return {};
  }
  try {
    return JSON.parse(text);
  } catch (err) {
    return {};
  }
}

async function apiRequest(path, options = {}) {
  const headers = Object.assign({ "Content-Type": "application/json" }, options.headers || {});
  if (state.token) {
    headers.Authorization = `Bearer ${state.token}`;
  }
  const response = await fetch(`${getBaseUrl()}${path}`, {
    ...options,
    headers
  });
  const data = await safeJson(response);
  if (!response.ok) {
    const message = data.error || response.statusText || "Request failed";
    throw new Error(message);
  }
  return data;
}

function renderAuth() {
  if (state.user) {
    sessionEl.hidden = false;
    sessionLabel.textContent = `Signed in as ${state.user.username}`;
  } else {
    sessionEl.hidden = true;
  }
  commentForm.querySelector("button").disabled = !state.user;
  savePostBtn.disabled = !state.user;
}

function formatDate(value) {
  if (!value) {
    return "";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "";
  }
  return new Intl.DateTimeFormat("en", {
    dateStyle: "medium",
    timeStyle: "short"
  }).format(date);
}

function resolveAuthor(post) {
  if (post.author && post.author.username) {
    return post.author.username;
  }
  if (post.user && post.user.username) {
    return post.user.username;
  }
  return `User ${post.user_id || "?"}`;
}

function canEditPost(post) {
  if (!state.user) {
    return false;
  }
  const authorId = post.user_id || (post.author ? post.author.id : null);
  return authorId === state.user.id;
}

function renderPosts(posts) {
  postsEl.innerHTML = "";
  if (!posts.length) {
    const empty = document.createElement("div");
    empty.className = "empty";
    empty.textContent = "No posts yet. Publish the first one.";
    postsEl.appendChild(empty);
    return;
  }

  posts.forEach((post) => {
    const card = document.createElement("article");
    card.className = "post-card";

    const title = document.createElement("h3");
    title.textContent = post.title;

    const meta = document.createElement("div");
    meta.className = "post-meta";
    const author = document.createElement("span");
    author.textContent = `By ${resolveAuthor(post)}`;
    const time = document.createElement("span");
    time.textContent = formatDate(post.created_at);
    meta.append(author, time);

    const body = document.createElement("p");
    body.textContent = post.content;

    const actions = document.createElement("div");
    actions.className = "row";

    const viewBtn = document.createElement("button");
    viewBtn.type = "button";
    viewBtn.className = "ghost";
    viewBtn.textContent = "View";
    viewBtn.addEventListener("click", () => selectPost(post));
    actions.appendChild(viewBtn);

    if (canEditPost(post)) {
      const editBtn = document.createElement("button");
      editBtn.type = "button";
      editBtn.className = "secondary";
      editBtn.textContent = "Edit";
      editBtn.addEventListener("click", () => startEdit(post));

      const deleteBtn = document.createElement("button");
      deleteBtn.type = "button";
      deleteBtn.textContent = "Delete";
      deleteBtn.addEventListener("click", () => deletePost(post));

      actions.append(editBtn, deleteBtn);
    }

    card.append(title, meta, body, actions);
    postsEl.appendChild(card);
  });
}

function renderComments(comments) {
  commentsEl.innerHTML = "";
  if (!comments.length) {
    const empty = document.createElement("div");
    empty.className = "empty";
    empty.textContent = "No comments yet. Start the discussion.";
    commentsEl.appendChild(empty);
    return;
  }

  comments.forEach((comment) => {
    const card = document.createElement("article");
    card.className = "post-card";
    const meta = document.createElement("div");
    meta.className = "post-meta";
    const author = document.createElement("span");
    const username = comment.user && comment.user.username ? comment.user.username : `User ${comment.user_id || "?"}`;
    author.textContent = `By ${username}`;
    const time = document.createElement("span");
    time.textContent = formatDate(comment.created_at);
    meta.append(author, time);

    const body = document.createElement("p");
    body.textContent = comment.content;

    card.append(meta, body);
    commentsEl.appendChild(card);
  });
}

async function loadPosts() {
  try {
    const posts = await apiRequest("/api/posts");
    state.posts = posts;
    renderPosts(posts);
    setStatus("Posts refreshed.", "success");
  } catch (err) {
    setStatus(err.message, "error");
  }
}

async function selectPost(post) {
  state.selectedPostId = post.id;
  postDetailsEl.textContent = `${post.title} — ${resolveAuthor(post)}`;
  commentSectionEl.hidden = false;
  await loadComments(post.id);
}

async function loadComments(postId) {
  try {
    const comments = await apiRequest(`/api/posts/${postId}/comments`);
    renderComments(comments);
  } catch (err) {
    setStatus(err.message, "error");
  }
}

function startEdit(post) {
  postForm.title.value = post.title;
  postForm.content.value = post.content;
  state.editingPostId = post.id;
  savePostBtn.textContent = "Update Post";
  cancelEditBtn.hidden = false;
}

function resetEdit() {
  postForm.reset();
  state.editingPostId = null;
  savePostBtn.textContent = "Publish Post";
  cancelEditBtn.hidden = true;
}

async function deletePost(post) {
  if (!confirm(`Delete "${post.title}"?`)) {
    return;
  }
  try {
    await apiRequest(`/api/posts/${post.id}`, { method: "DELETE" });
    setStatus("Post deleted.", "success");
    if (state.selectedPostId === post.id) {
      postDetailsEl.textContent = "Select a post to view comments and reply.";
      commentSectionEl.hidden = true;
      commentsEl.innerHTML = "";
    }
    await loadPosts();
  } catch (err) {
    setStatus(err.message, "error");
  }
}

registerForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  const payload = Object.fromEntries(new FormData(registerForm).entries());
  try {
    await apiRequest("/api/register", {
      method: "POST",
      body: JSON.stringify(payload)
    });
    setStatus("Registered successfully. Please sign in.", "success");
    registerForm.reset();
  } catch (err) {
    setStatus(err.message, "error");
  }
});

loginForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  const payload = Object.fromEntries(new FormData(loginForm).entries());
  try {
    const data = await apiRequest("/api/login", {
      method: "POST",
      body: JSON.stringify(payload)
    });
    state.token = data.token;
    state.user = data.user;
    saveAuth();
    renderAuth();
    setStatus("Signed in successfully.", "success");
    loginForm.reset();
    await loadPosts();
  } catch (err) {
    setStatus(err.message, "error");
  }
});

logoutBtn.addEventListener("click", () => {
  clearAuth();
  renderAuth();
  setStatus("Signed out.", "success");
});

postForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  if (!state.user) {
    setStatus("Sign in to publish posts.", "error");
    return;
  }
  const payload = Object.fromEntries(new FormData(postForm).entries());
  try {
    if (state.editingPostId) {
      await apiRequest(`/api/posts/${state.editingPostId}`, {
        method: "PUT",
        body: JSON.stringify(payload)
      });
      setStatus("Post updated.", "success");
    } else {
      await apiRequest("/api/posts", {
        method: "POST",
        body: JSON.stringify(payload)
      });
      setStatus("Post published.", "success");
    }
    resetEdit();
    await loadPosts();
  } catch (err) {
    setStatus(err.message, "error");
  }
});

cancelEditBtn.addEventListener("click", () => {
  resetEdit();
});

refreshPostsBtn.addEventListener("click", () => {
  loadPosts();
});

commentForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  if (!state.user) {
    setStatus("Sign in to comment.", "error");
    return;
  }
  if (!state.selectedPostId) {
    setStatus("Select a post before commenting.", "error");
    return;
  }
  const payload = Object.fromEntries(new FormData(commentForm).entries());
  try {
    await apiRequest(`/api/posts/${state.selectedPostId}/comments`, {
      method: "POST",
      body: JSON.stringify(payload)
    });
    commentForm.reset();
    setStatus("Comment added.", "success");
    await loadComments(state.selectedPostId);
  } catch (err) {
    setStatus(err.message, "error");
  }
});

loadAuth();
renderAuth();
loadPosts();
