// Runtime configuration for Admin FE.
// Edit apiBaseUrl if you need to point the UI to a different tea-api instance
// without rebuilding the frontend bundle.
window.__TEA_RUNTIME_CONFIG__ = window.__TEA_RUNTIME_CONFIG__ || {};
window.__TEA_RUNTIME_CONFIG__.apiBaseUrl =
  window.__TEA_RUNTIME_CONFIG__.apiBaseUrl || 'http://127.0.0.1:9292';
