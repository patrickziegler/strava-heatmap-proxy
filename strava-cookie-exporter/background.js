chrome.tabs.onUpdated.addListener((tabId, changeInfo, tab) => {
    if (changeInfo.status === 'complete' && tab.url.includes('strava.com')) {
        chrome.runtime.sendMessage({ action: 'pageLoadComplete' });
    }
});

chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
    if (message.action === 'exportStravaCookies') {
        exportToFile(message.cookies);
    }
});

let downloadMap = new Map();

async function exportToFile(data, filename = 'strava-cookies.json') {
    const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
    const blobUrl = URL.createObjectURL(blob);

    chrome.downloads.download({
        url: blobUrl,
        filename: filename,
        saveAs: true,
    }, (downloadId) => {
        if (chrome.runtime.lastError) {
            console.error("Download failed:", chrome.runtime.lastError);
            URL.revokeObjectURL(blobUrl);
            return;
        } else {
            downloadMap.set(downloadId, blobUrl);
        }
    });

    chrome.downloads.onChanged.addListener(function listener(download) {
        const blobUrl = downloadMap.get(download.id);
        if (!blobUrl) {
            return; // download id not found in map
        }
        if (download.state && download.state.current === 'complete') {
            URL.revokeObjectURL(blobUrl);
            downloadMap.delete(download.id);
            chrome.downloads.onChanged.removeListener(listener);
        }
    });
}
