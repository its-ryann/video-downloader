async function startDownload() {
  const url = document.getElementById("urlInput").value.trim();
  const format = document.getElementById("formatSelect").value;
  const status = document.getElementById("status");
  const btn = document.getElementById("downloadBtn");

  // Validate input
  if (!url) {
    status.textContent = "Please enter a video URL.";
    status.className = "status error";
    return;
  }

  // Show loading state
  btn.disabled = true;
  btn.textContent = "Downloading...";
  status.textContent = "Please wait, this may take a minute...";
  status.className = "status info";

  try {
    const response = await fetch("/download", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ url, format }),
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(errorText);
    }

    // Trigger file download in the browser
    const blob = await response.blob();
    const downloadUrl = window.URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = downloadUrl;
    a.download = `video.${format}`;
    document.body.appendChild(a);
    a.click();
    a.remove();
    window.URL.revokeObjectURL(downloadUrl);

    status.textContent = "Download complete!";
    status.className = "status success";

  } catch (err) {
    status.textContent = "Error: " + err.message;
    status.className = "status error";

  } finally {
    btn.disabled = false;
    btn.textContent = "Download";
  }
}