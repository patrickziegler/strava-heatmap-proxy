chrome.tabs.onUpdated.addListener((tabId, changeInfo, tab) => {
    if (changeInfo.status === 'complete' && tab && tab.url && tab.url.includes('strava.com')) {
        chrome.runtime.sendMessage({ action: 'pageLoadComplete' });
    }
});

chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
    if (message.action === 'exportStravaCookies') {
        exportToFile(message.cookies);
    }
});

async function exportToFile(data, filename = 'strava-cookies.json') {
    // Convert data to JSON string and create data URL
    const jsonString = JSON.stringify(data, null, 2);
    const dataUrl = 'data:application/json;charset=utf-8,' + encodeURIComponent(jsonString);

    chrome.downloads.download({
        url: dataUrl,
        filename: filename,
        saveAs: true,
    }, (downloadId) => {
        if (chrome.runtime.lastError) {
            console.error("Download failed:", chrome.runtime.lastError);
        }
    });
}
