chrome.tabs.query({
  active: true,
  currentWindow: true
}, function(tabs) {
  chrome.cookies.getAll({}, function (cookies) {
    var cookieAuth = {};
    for (var i in cookies) {
      var cookie = cookies[i];
      cookieAuth[cookie.name] = cookie.value;
    }
    var str = JSON.stringify(cookieAuth, null, 2);
    var el = document.createElement('textarea');
    // Set value (string to be copied)
    el.value = str;
    // Set non-editable to avoid focus and move outside of view
    el.setAttribute('readonly', '');
    el.style = {position: 'absolute', left: '-9999px'};
    document.body.appendChild(el);
    // Select text inside element
    el.select();
    // Copy text to clipboard
    document.execCommand('copy');
    // Remove temporary element
    document.body.removeChild(el);
  });
});