let activeInterval = null;
let activeJobId = null;
let currentVideoInfo = null;
let debounceTimeout = null;

document.addEventListener("DOMContentLoaded", () => {
  const urlInput = document.getElementById("urlInput");
  if (urlInput) {
    urlInput.addEventListener("input", () => {
      clearTimeout(debounceTimeout);
      debounceTimeout = setTimeout(() => {
        const url = urlInput.value.trim();
        if (url.startsWith("http://") || url.startsWith("https://")) {
          fetchMetadata();
        } else {
          hidePreview();
        }
      }, 600);
    });
  }
  loadSettings();
  renderLibrary();
});

// View Navigation Logic
function switchTab(tabName) {
  const views = {
    home: document.getElementById("viewHome"),
    library: document.getElementById("viewLibrary"),
    settings: document.getElementById("viewSettings")
  };

  const navs = {
    home: document.getElementById("navHome"),
    library: document.getElementById("navLibrary"),
    settings: document.getElementById("navSettings")
  };

  const mobileNavs = {
    home: document.getElementById("mobileNavHome"),
    library: document.getElementById("mobileNavLibrary"),
    settings: document.getElementById("mobileNavSettings")
  };

  Object.keys(views).forEach(key => {
    if (views[key]) {
      if (key === tabName) {
        views[key].classList.remove("hidden");
      } else {
        views[key].classList.add("hidden");
      }
    }

    if (navs[key]) {
      if (key === tabName) {
        navs[key].classList.add("text-white", "active");
        navs[key].classList.remove("text-on-surface-variant");
      } else {
        navs[key].classList.remove("text-white", "active");
        navs[key].classList.add("text-on-surface-variant");
      }
    }

    if (mobileNavs[key]) {
      if (key === tabName) {
        mobileNavs[key].classList.add("text-white");
        mobileNavs[key].classList.remove("text-on-surface-variant/70");
      } else {
        mobileNavs[key].classList.remove("text-white");
        mobileNavs[key].classList.add("text-on-surface-variant/70");
      }
    }
  });

  if (tabName === "library") {
    renderLibrary();
  }
  if (tabName === "settings") {
    updateCacheSizeDisplay();
  }
}

function formatDuration(seconds) {
  if (!seconds || isNaN(seconds)) return "Live";
  const mins = Math.floor(seconds / 60);
  const secs = Math.floor(seconds % 60);
  return `${mins}:${secs < 10 ? "0" : ""}${secs}`;
}

function toggleQualitySelect() {
  const format = document.getElementById("formatSelect").value;
  const qualityGroup = document.getElementById("qualityGroup");
  if (format === "mp3") {
    qualityGroup.style.opacity = "0.4";
    qualityGroup.style.pointerEvents = "none";
  } else {
    qualityGroup.style.opacity = "1";
    qualityGroup.style.pointerEvents = "auto";
  }
}

async function fetchMetadata() {
  const url = document.getElementById("urlInput").value.trim();
  const fetchBtnText = document.getElementById("fetchBtnText");

  if (!url) return;

  fetchBtnText.textContent = "Loading...";

  try {
    const res = await fetch("/info", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ url }),
    });

    if (!res.ok) throw new Error("Could not fetch video preview.");

    const info = await res.json();
    currentVideoInfo = info;

    document.getElementById("previewThumb").src = info.thumbnail || "";
    document.getElementById("previewTitle").textContent = info.title || "Video Preview";
    document.getElementById("previewUploader").textContent = info.uploader ? `By ${info.uploader}` : "";
    document.getElementById("previewDuration").textContent = formatDuration(info.duration);
    document.getElementById("previewPlatform").textContent = info.platform || "Video";

    document.getElementById("previewCard").classList.remove("hidden");
  } catch (err) {
    hidePreview();
  } finally {
    fetchBtnText.textContent = "Preview";
  }
}

function hidePreview() {
  currentVideoInfo = null;
  document.getElementById("previewCard").classList.add("hidden");
}

async function startDownload() {
  const url = document.getElementById("urlInput").value.trim();
  const format = document.getElementById("formatSelect").value;
  const quality = document.getElementById("qualitySelect").value;
  const btn = document.getElementById("downloadBtn");
  const btnText = document.getElementById("btnText");

  if (!url) {
    showStatus("error", "Please paste a valid video URL first.", false);
    return;
  }

  btn.disabled = true;
  btnText.textContent = "Connecting...";
  document.getElementById("cancelBtn").classList.remove("hidden");
  showStatus("", "Initializing download process...", true, 0);

  const title = currentVideoInfo ? currentVideoInfo.title : "";

  try {
    const res = await fetch("/download", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ url, format, quality, title }),
    });

    if (!res.ok) throw new Error(await res.text());
    const { job_id } = await res.json();
    activeJobId = job_id;
    await pollProgress(job_id, format, quality, url);
  } catch (err) {
    showStatus("error", "Error: " + err.message, false);
    btn.disabled = false;
    btnText.textContent = "Download Now";
    document.getElementById("cancelBtn").classList.add("hidden");
  }
}

async function pollProgress(jobId, format, quality, originalUrl) {
  return new Promise((resolve, reject) => {
    activeInterval = setInterval(async () => {
      try {
        const res = await fetch(`/progress/${jobId}`);
        if (!res.ok) {
          throw new Error("Job status check failed.");
        }
        const data = await res.json();

        if (data.status === "processing") {
          const pct = data.progress || 0;
          showStatus("", `Downloading media...`, true, pct);
          document.getElementById("btnText").textContent = `Downloading ${pct}%`;
        }

        if (data.status === "cancelled") {
          cleanupPolling();
          showStatus("error", "Download cancelled by user.", false);
          resetButtons();
          resolve();
        }

        if (data.status === "error") {
          cleanupPolling();
          showStatus("error", "Error: " + (data.error || "Failed to download video"), false);
          resetButtons();
          reject(new Error(data.error));
        }

        if (data.status === "done") {
          cleanupPolling();
          showStatus("", "Preparing file download...", true, 100);

          const fileRes = await fetch(`/file/${jobId}?dl=1`);
          if (!fileRes.ok) throw new Error(await fileRes.text());

          const blob = await fileRes.blob();
          const downloadUrl = window.URL.createObjectURL(blob);
          const a = document.createElement("a");
          a.href = downloadUrl;

          const downloadFilename = data.filename || `vidsnap_${jobId}.${format}`;
          a.download = downloadFilename;

          document.body.appendChild(a);
          a.click();
          a.remove();
          window.URL.revokeObjectURL(downloadUrl);

          showStatus("success", `Download ready: ${downloadFilename}`, false);
          
          const settings = getSettings();
          if (settings.autoSave !== false) {
            saveToLibrary({
              id: jobId,
              originalUrl: originalUrl || "",
              title: (currentVideoInfo && currentVideoInfo.title) || downloadFilename,
              uploader: (currentVideoInfo && currentVideoInfo.uploader) || "VidSnap Download",
              thumbnail: (currentVideoInfo && currentVideoInfo.thumbnail) || "https://images.unsplash.com/photo-1618005182384-a83a8bd57fbe?q=80&w=600&auto=format&fit=crop",
              duration: (currentVideoInfo && currentVideoInfo.duration) ? formatDuration(currentVideoInfo.duration) : "00:00",
              format: format.toUpperCase(),
              quality: quality || "HD",
              platform: (currentVideoInfo && currentVideoInfo.platform) || "Web",
              date: new Date().toLocaleDateString()
            });
          }

          document.getElementById("urlInput").value = "";
          hidePreview();
          resetButtons();
          resolve();
        }
      } catch (err) {
        cleanupPolling();
        showStatus("error", "Error: " + err.message, false);
        resetButtons();
        reject(err);
      }
    }, 800);
  });
}

function cleanupPolling() {
  if (activeInterval) {
    clearInterval(activeInterval);
    activeInterval = null;
  }
}

function resetButtons() {
  const btn = document.getElementById("downloadBtn");
  btn.disabled = false;
  document.getElementById("btnText").textContent = "Download Now";
  document.getElementById("cancelBtn").classList.add("hidden");
  activeJobId = null;
}

async function cancelDownload() {
  if (activeJobId) {
    try {
      await fetch(`/cancel/${activeJobId}`, { method: "POST" });
    } catch (e) {
      console.warn("Cancel request failed", e);
    }
  }
  cleanupPolling();
  resetButtons();
  showStatus("error", "Download cancelled.", false);
}

function showStatus(type, message, showSpinner, progress) {
  const statusBox = document.getElementById("statusBox");
  const statusText = document.getElementById("statusText");
  const spinner = document.getElementById("spinner");
  const progressWrap = document.getElementById("progressBarWrap");
  const progressBar = document.getElementById("progressBar");
  const progressPct = document.getElementById("progressPct");

  statusBox.className = "flex flex-col gap-3 p-4 rounded-xl bg-black/40 border " + (type === "error" ? "border-error/40" : "border-white/10");
  statusText.textContent = message;
  spinner.className = "w-4 h-4 border-2 border-white/20 border-t-white rounded-full animate-spin" + (showSpinner ? "" : " hidden");

  if (progress !== undefined && progress >= 0) {
    statusBox.classList.remove("hidden");
    progressWrap.classList.remove("hidden");
    progressBar.style.width = progress + "%";
    progressPct.textContent = progress + "%";
    progressPct.style.display = "inline";
  } else {
    if (!message) {
      statusBox.classList.add("hidden");
    } else {
      statusBox.classList.remove("hidden");
    }
    progressWrap.classList.add("hidden");
    progressBar.style.width = "0%";
    progressPct.style.display = "none";
  }
}

// MEDIA PLAYER MODAL
function openPlayerModal(id) {
  const lib = getLibrary();
  const item = lib.find(i => i.id === id);
  if (!item) return;

  const modal = document.getElementById("playerModal");
  const titleEl = document.getElementById("modalTitle");
  const platformEl = document.getElementById("modalPlatform");
  const formatEl = document.getElementById("modalFormat");
  const qualityEl = document.getElementById("modalQuality");
  const downloadLink = document.getElementById("modalDownloadLink");
  const videoPlayer = document.getElementById("modalVideoPlayer");
  const audioWrap = document.getElementById("modalAudioPlayerWrap");
  const audioPlayer = document.getElementById("modalAudioPlayer");
  const audioThumb = document.getElementById("modalAudioThumb");

  titleEl.textContent = item.title;
  platformEl.textContent = item.platform;
  formatEl.textContent = item.format;
  qualityEl.textContent = item.quality;
  downloadLink.href = `/file/${item.id}?dl=1`;

  const fileUrl = `/file/${item.id}`;

  if (item.format === "MP3") {
    videoPlayer.classList.add("hidden");
    videoPlayer.pause();
    videoPlayer.src = "";

    audioWrap.classList.remove("hidden");
    audioThumb.src = item.thumbnail;
    audioPlayer.src = fileUrl;
    audioPlayer.play().catch(() => {});
  } else {
    audioWrap.classList.add("hidden");
    audioPlayer.pause();
    audioPlayer.src = "";

    videoPlayer.classList.remove("hidden");
    videoPlayer.src = fileUrl;
    videoPlayer.play().catch(() => {});
  }

  modal.classList.remove("hidden");
}

function closePlayerModal() {
  const modal = document.getElementById("playerModal");
  const videoPlayer = document.getElementById("modalVideoPlayer");
  const audioPlayer = document.getElementById("modalAudioPlayer");

  if (videoPlayer) {
    videoPlayer.pause();
    videoPlayer.src = "";
  }
  if (audioPlayer) {
    audioPlayer.pause();
    audioPlayer.src = "";
  }
  if (modal) {
    modal.classList.add("hidden");
  }
}

// SETTINGS MANAGEMENT
function getSettings() {
  try {
    return JSON.parse(localStorage.getItem("vidsnap_settings")) || {
      defaultFormat: "mp4",
      defaultQuality: "720p",
      autoSave: true,
      liveProgress: true
    };
  } catch (e) {
    return { defaultFormat: "mp4", defaultQuality: "720p", autoSave: true, liveProgress: true };
  }
}

function loadSettings() {
  const settings = getSettings();
  const fmtEl = document.getElementById("settingDefaultFormat");
  const qualEl = document.getElementById("settingDefaultQuality");
  const autoSaveEl = document.getElementById("settingAutoSave");
  const progressEl = document.getElementById("settingLiveProgress");

  if (fmtEl) fmtEl.value = settings.defaultFormat || "mp4";
  if (qualEl) qualEl.value = settings.defaultQuality || "720p";
  if (autoSaveEl) autoSaveEl.checked = settings.autoSave !== false;
  if (progressEl) progressEl.checked = settings.liveProgress !== false;

  const homeFormat = document.getElementById("formatSelect");
  const homeQuality = document.getElementById("qualitySelect");
  if (homeFormat) homeFormat.value = settings.defaultFormat || "mp4";
  if (homeQuality) homeQuality.value = settings.defaultQuality || "720p";
  toggleQualitySelect();

  updateCacheSizeDisplay();
}

function saveSettings() {
  const settings = {
    defaultFormat: document.getElementById("settingDefaultFormat").value,
    defaultQuality: document.getElementById("settingDefaultQuality").value,
    autoSave: document.getElementById("settingAutoSave").checked,
    liveProgress: document.getElementById("settingLiveProgress").checked
  };

  localStorage.setItem("vidsnap_settings", JSON.stringify(settings));

  const homeFormat = document.getElementById("formatSelect");
  const homeQuality = document.getElementById("qualitySelect");
  if (homeFormat) homeFormat.value = settings.defaultFormat;
  if (homeQuality) homeQuality.value = settings.defaultQuality;
  toggleQualitySelect();
}

function updateCacheSizeDisplay() {
  const cacheEl = document.getElementById("cacheSizeText");
  if (!cacheEl) return;
  const items = getLibrary();
  cacheEl.textContent = `Items Saved: ${items.length}`;
}

// LIBRARY HISTORY PERSISTENCE
function getLibrary() {
  try {
    return JSON.parse(localStorage.getItem("vidsnap_library")) || [];
  } catch (e) {
    return [];
  }
}

function saveToLibrary(item) {
  const lib = getLibrary();
  lib.unshift(item);
  localStorage.setItem("vidsnap_library", JSON.stringify(lib.slice(0, 30)));
  updateCacheSizeDisplay();
}

function clearLibraryHistory() {
  localStorage.removeItem("vidsnap_library");
  renderLibrary();
  updateCacheSizeDisplay();
}

function filterLibrary() {
  const query = document.getElementById("librarySearch").value.toLowerCase().trim();
  renderLibrary(query);
}

function renderLibrary(filterQuery = "") {
  const libraryGrid = document.getElementById("libraryGrid");
  const libraryEmpty = document.getElementById("libraryEmpty");
  if (!libraryGrid) return;

  const items = getLibrary().filter(item => {
    if (!filterQuery) return true;
    return item.title.toLowerCase().includes(filterQuery) || item.platform.toLowerCase().includes(filterQuery);
  });

  if (items.length === 0) {
    libraryGrid.innerHTML = "";
    libraryEmpty.classList.remove("hidden");
    return;
  }

  libraryEmpty.classList.add("hidden");
  libraryGrid.innerHTML = items.map(item => `
    <div onclick="openPlayerModal('${item.id}')" class="glass-panel rounded-2xl overflow-hidden flex flex-col group cursor-pointer hover:border-white/40 hover:scale-[1.02] transition-all">
      <div class="relative aspect-video overflow-hidden bg-black/40">
        <img class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500" src="${item.thumbnail}" alt="${item.title}"/>
        <div class="absolute inset-0 bg-gradient-to-t from-black/80 via-transparent to-transparent opacity-80"></div>
        <div class="absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity bg-black/40">
          <div class="w-12 h-12 rounded-full bg-white text-black flex items-center justify-center shadow-lg">
            <span class="material-symbols-outlined text-2xl" style="font-variation-settings: 'FILL' 1;">play_arrow</span>
          </div>
        </div>
        <div class="absolute bottom-2.5 right-2.5 bg-black/80 backdrop-blur-md text-white font-mono text-[11px] px-2 py-0.5 rounded border border-white/10">
          ${item.duration}
        </div>
        <div class="absolute top-2.5 left-2.5 bg-white/10 backdrop-blur-md text-white text-[10px] font-bold px-2 py-0.5 rounded-full border border-white/20">
          ${item.platform}
        </div>
      </div>
      <div class="p-4 flex flex-col gap-2">
        <h3 class="font-bold text-white text-base truncate">${item.title}</h3>
        <div class="flex items-center justify-between text-xs text-on-surface-variant mt-1">
          <div class="flex gap-1.5">
            <span class="px-2 py-0.5 rounded bg-white/10 text-white text-[10px] font-semibold">${item.format}</span>
            <span class="px-2 py-0.5 rounded bg-white/10 text-white text-[10px] font-semibold">${item.quality}</span>
          </div>
          <span class="text-[11px]">${item.date}</span>
        </div>
      </div>
    </div>
  `).join("");
}



