
// Import web components
import './wc/root'
import './wc/core/Provider'
import './wc/core/Array'
import './wc/layout/Table'
import './wc/core/GoogleAuth'

// Import classes
import { GoogleAuth } from './wc/core/GoogleAuth';
import { Toast } from './wc/layout/Toast';
import { Button } from './wc/layout/Button'

// Initialize the app
document.addEventListener("DOMContentLoaded", () => {
    const googleAuth = document.getElementById('auth') as GoogleAuth;
    googleAuth.opts = {
        ClientId: '121760808688-hbiibnih1trt2vrokhrta17jgeuagp4k.apps.googleusercontent.com',
        Theme: 'outline',
        Size: 'small'
    };


    const toast = document.querySelector('wc-toast') as Toast;
    const buttons = document.querySelectorAll('wc-button');
    buttons.forEach((button) => {
        button.addEventListener('click', (evt) => {
            const button = evt.target as Button;
            toast.show(`Button clicked ${ button.textContent }`,{
                duration: 3000,
                color: button.color
            });
        });
    });
});
