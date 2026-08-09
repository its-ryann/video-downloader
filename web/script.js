let activeInterval = null;
let activeJobId = null;
let currentVideoInfo = null;
let debounceTimeout = null;

document.addEventListener("DOMContentLoaded", () => {
  const urlInput = document.getElementById("urlInput");
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
});

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
    await pollProgress(job_id, format);
  } catch (err) {
    showStatus("error", "Error: " + err.message, false);
    btn.disabled = false;
    btnText.textContent = "Download Now";
    document.getElementById("cancelBtn").classList.add("hidden");
  }
}

async function pollProgress(jobId, format) {
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

          const fileRes = await fetch(`/file/${jobId}`);
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

  statusBox.className = "status-box" + (type ? " " + type : "");
  statusText.textContent = message;
  spinner.className = "spinner" + (showSpinner ? "" : " hidden");

  if (progress !== undefined && progress >= 0) {
    progressWrap.classList.remove("hidden");
    progressBar.style.width = progress + "%";
    progressPct.textContent = progress + "%";
    progressPct.style.display = "inline";
  } else {
    progressWrap.classList.add("hidden");
    progressBar.style.width = "0%";
    progressPct.style.display = "none";
  }
}

