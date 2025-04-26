
// Import web components
import './wc/GoogleAuth'
import './wc/Provider'
import './wc/Array'

// Import classes
import { GoogleAuth } from './wc/GoogleAuth';

// Initialize the app
document.addEventListener("DOMContentLoaded", () => {
    const googleAuth = document.getElementById('auth') as GoogleAuth;
    googleAuth.opts = {
        ClientId: '121760808688-hbiibnih1trt2vrokhrta17jgeuagp4k.apps.googleusercontent.com',
        Theme: 'outline',
        Size: 'small'
    };
});
