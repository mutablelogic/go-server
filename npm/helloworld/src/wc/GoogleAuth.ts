
export class GoogleAuth extends Object {
  constructor(client_id: string) {
    super();

    window['google'].accounts.id.initialize({
      client_id: client_id,
      callback: this.credentialResponse
    });
  }

  render(parent: HTMLElement) {
    window['google'].accounts.id.renderButton(
      parent, { theme: 'outline', size: 'large' }
    );
  }

  private credentialResponse(response: any) {
    // This function is called when the user successfully logs in
    console.log("Encoded JWT ID token: " + response.credential);
  }  
}
