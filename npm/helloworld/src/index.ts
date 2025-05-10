
// Import web components
import './wc/root'
import './wc/core/Provider'
import './wc/core/Array'
import './wc/layout/Table'
import './wc/core/GoogleAuth'

// Import classes
import { GoogleAuth } from './wc/core/GoogleAuth';

// Initialize the app
document.addEventListener("DOMContentLoaded", () => {
    const googleAuth = document.getElementById('auth') as GoogleAuth;
    googleAuth.opts = {
        ClientId: '121760808688-hbiibnih1trt2vrokhrta17jgeuagp4k.apps.googleusercontent.com',
        Theme: 'outline',
        Size: 'small'
    };
});
