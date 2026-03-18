// Toast Manager
function showToast(message, type) {
    type = type || 'success';
    var bgClass = type === 'error' ? 'text-bg-danger' : 'text-bg-' + type;
    var iconClass = type === 'success' ? 'df-icon-check-circle' : 'df-icon-warning';

    var toastEl = document.createElement('div');
    toastEl.className = 'toast align-items-center ' + bgClass + ' border-0';
    toastEl.setAttribute('role', 'alert');

    var wrapper = document.createElement('div');
    wrapper.className = 'd-flex';

    var body = document.createElement('div');
    body.className = 'toast-body';

    var icon = document.createElement('i');
    icon.className = iconClass + ' me-1';
    body.appendChild(icon);
    body.appendChild(document.createTextNode(' ' + message));

    var closeBtn = document.createElement('button');
    closeBtn.type = 'button';
    closeBtn.className = 'btn-close btn-close-white me-2 m-auto';
    closeBtn.setAttribute('data-bs-dismiss', 'toast');

    wrapper.appendChild(body);
    wrapper.appendChild(closeBtn);
    toastEl.appendChild(wrapper);

    document.getElementById('toast-container').appendChild(toastEl);
    var toast = new bootstrap.Toast(toastEl, { delay: 4000 });
    toast.show();

    toastEl.addEventListener('hidden.bs.toast', function() {
        toastEl.remove();
    });
}

// Copy to clipboard
function copyToClipboard(text, btn) {
    navigator.clipboard.writeText(text).then(function() {
        btn.classList.add('copied');
        var origIcon = btn.querySelector('i');
        var origClass = origIcon ? origIcon.className : '';
        if (origIcon) {
            origIcon.className = 'df-icon-check';
        }
        setTimeout(function() {
            if (origIcon) {
                origIcon.className = origClass;
            }
            btn.classList.remove('copied');
        }, 2000);
        showToast('Copied to clipboard!', 'success');
    });
}

// Download QR code as proper file
function downloadQR(dataURI, filename) {
    // Parse the data URI to get mime type and data
    var parts = dataURI.split(',');
    var meta = parts[0]; // e.g. "data:image/svg+xml;base64"
    var data = parts[1];
    var mime = meta.split(':')[1].split(';')[0];

    // Determine file extension from mime type
    var ext = 'png';
    if (mime.indexOf('svg') !== -1) ext = 'svg';
    else if (mime.indexOf('png') !== -1) ext = 'png';
    else if (mime.indexOf('jpeg') !== -1 || mime.indexOf('jpg') !== -1) ext = 'jpg';

    // Decode base64 to binary
    var byteString = atob(data);
    var ab = new ArrayBuffer(byteString.length);
    var ia = new Uint8Array(ab);
    for (var i = 0; i < byteString.length; i++) {
        ia[i] = byteString.charCodeAt(i);
    }

    var blob = new Blob([ab], { type: mime });
    var url = URL.createObjectURL(blob);

    var a = document.createElement('a');
    a.href = url;
    a.download = filename + '.' + ext;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
}

// Initialize tooltips on page load
document.addEventListener('DOMContentLoaded', function() {
    var tooltipTriggerList = [].slice.call(document.querySelectorAll('[data-bs-toggle="tooltip"]'));
    tooltipTriggerList.map(function(el) { return new bootstrap.Tooltip(el); });

    // Mobile sidebar overlay toggle
    var sidebar = document.getElementById('sidebar');
    var overlay = document.querySelector('.sidebar-overlay');
    if (sidebar && overlay) {
        var observer = new MutationObserver(function() {
            overlay.style.display = sidebar.classList.contains('show') ? 'block' : 'none';
        });
        observer.observe(sidebar, { attributes: true, attributeFilter: ['class'] });
    }
});
