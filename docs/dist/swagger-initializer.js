window.onload = function() {
  //<editor-fold desc="Changeable Configuration Block">
  const queryString = window.location.search;
  const urlParams = new URLSearchParams(queryString);
  const swaggerUrl = urlParams.get('url')
  // the following lines will be replaced by docker/configurator, when it runs in a docker-container
  window.ui = SwaggerUIBundle({
    url: swaggerUrl,
    dom_id: '#swagger-ui',
    deepLinking: true,
    presets: [
      SwaggerUIBundle.presets.apis,
      SwaggerUIStandalonePreset
    ],
    plugins: [
      SwaggerUIBundle.plugins.DownloadUrl
    ],
    layout: "StandaloneLayout"
  });

  //</editor-fold>
};
