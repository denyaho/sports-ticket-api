const { URL } = require('url');

function tryParseHttpUrl (url) {
  try {
    const parsedUrl = new URL(url);

    if (parsedUrl.protocol !== 'http:' && parsedUrl.protocol !== 'https:') {
      return { isSuccess: false, data: 'URL protocol must be HTTP or HTTPS' };
    }

    // Remove trailing slash from the origin.
    const sanitizedOrigin = parsedUrl.origin.replace(/\/$/, '');

    return { isSuccess: true, data: sanitizedOrigin };
  } catch (error) {
    return { isSuccess: false, data: 'Invalid URL format' };
  }
}

module.exports = {
  tryParseHttpUrl,
};
