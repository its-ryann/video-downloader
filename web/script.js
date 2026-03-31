async function startDownload() {
  const url = document.getElementById("urlInput").value.trim();
  const format = document.getElementById("formatSelect").value;
  const btn = document.getElementById("downloadBtn");
  const btnText = document.getElementById("btnText");

  if (!url) {
    showStatus("error", "Please paste a video URL first.", false);
    return;
  }

  btn.disabled = true;
  btnText.textContent = "Starting...";
  document.getElementById("cancelBtn").classList.remove("hidden");

  try {
    // Step 1: Start the download job
    const res = await fetch("/download", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ url, format }),
    });

    if (!res.ok) throw new Error(await res.text());

    const { job_id } = await res.json();

    // Step 2: Poll progress every second
    await pollProgress(job_id, format);

  } catch (err) {
    showStatus("error", "Error: " + err.message, false);
  } finally {
    btn.disabled = false;
    btnText.textContent = "Download";
    document.getElementById("cancelBtn").classList.add("hidden");
  }
}

async function pollProgress(jobId, format) {
  return new Promise((resolve, reject) => {
    activeInterval = setInterval(async () => {
      try {
        const res = await fetch(`/progress/${jobId}`);
        const data = await res.json();

        if (data.status === "processing") {
          const pct = data.progress;
          showStatus("", `Downloading... ${pct}%`, true, pct);
          document.getElementById("btnText").textContent = `${pct}%`;
        }

        if (data.status === "error") {
          clearInterval(interval);
          showStatus("error", "Error: " + data.error, false);
          reject(new Error(data.error));
        }

        if (data.status === "done") {
          clearInterval(interval);
          showStatus("", "Preparing file...", true, 100);

          // Step 3: Fetch the file
          const fileRes = await fetch(`/file/${jobId}`);
          if (!fileRes.ok) throw new Error(await fileRes.text());

          const blob = await fileRes.blob();
          const downloadUrl = window.URL.createObjectURL(blob);
          const a = document.createElement("a");
          a.href = downloadUrl;
          a.download = `video.${format}`;
          document.body.appendChild(a);
          a.click();
          a.remove();
          window.URL.revokeObjectURL(downloadUrl);

          showStatus("success", "Download complete!", false);
          document.getElementById("urlInput").value = "";
          resolve();
        }

      } catch (err) {
        clearInterval(interval);
        showStatus("error", "Error: " + err.message, false);
        reject(err);
      }
    }, 1000);
  });
}

function showStatus(type, message, showSpinner, progress) {
  const statusBox = document.getElementById("statusBox");
  const statusText = document.getElementById("statusText");
  const spinner = document.getElementById("spinner");
  const progressWrap = document.getElementById("progressBarWrap");
  const progressBar = document.getElementById("progressBar");

  statusBox.className = "status-box" + (type ? " " + type : "");
  statusText.textContent = message;
  spinner.className = "spinner" + (showSpinner ? "" : " hidden");

  if (progress !== undefined) {
    progressWrap.classList.remove("hidden");
    progressBar.style.width = progress + "%";
  } else {
    progressWrap.classList.add("hidden");
    progressBar.style.width = "0%";
  }
}

function cancelDownload() {
  if (activeInterval) {
    clearInterval(activeInterval);
    activeInterval = null;
  }
  document.getElementById("downloadBtn").disabled = false;
  document.getElementById("btnText").textContent = "Download";
  document.getElementById("cancelBtn").classList.add("hidden");
  showStatus("error", "Download cancelled.", false);
}
