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

function isFirefox() {
    return navigator.userAgent.toLowerCase().indexOf("firefox") > -1;
}

async function exportToFile(data, filename = 'strava-cookies.json') {
    const jsonString = JSON.stringify(data, null, 2);
    if (isFirefox()) {
        const blob = new Blob([jsonString], { type: 'application/json' });
        const url = URL.createObjectURL(blob);
        chrome.downloads.download({
            url: url,
            filename: filename,
            saveAs: true,
        }, (downloadId) => {
            URL.revokeObjectURL(url);
            if (chrome.runtime.lastError) {
                console.error("Download failed:", chrome.runtime.lastError);
            }
        });
    } else {
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
}
