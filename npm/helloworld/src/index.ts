
// Import web components
import './wc/App.ts'
import './wc/GoogleAuth.js'
import './wc/Provider.js'

// Import classes
import { GoogleAuth } from './wc/GoogleAuth.js';

// Live reload for esbuild
document.addEventListener("DOMContentLoaded", () => {
    new EventSource('/esbuild').addEventListener('change', () => location.reload())
});

// Initialize the app
document.addEventListener("DOMContentLoaded", () => {
    const googleAuth = document.getElementById('auth') as GoogleAuth;
    googleAuth.opts = {
        ClientId: '121760808688-hbiibnih1trt2vrokhrta17jgeuagp4k.apps.googleusercontent.com',
        Theme: 'outline',
        Size: 'large'
    };
});
