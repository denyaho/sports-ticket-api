const { writeFile } = require('fs/promises');

/* eslint-disable no-console */
function writeToConsole (output) {
  console.log(output);
}

async function writeToFileAsync (filePath, output) {
  if (typeof filePath !== 'string' || !filePath || !filePath.trim()) {
    throw new Error('File path must be a non-empty string.');
  }

  try {
    await writeFile(filePath, output);
  } catch (error) {
    const message = error.message ? error.message : `Unknown error when writing to ${filePath}`;

    throw new Error(message);
  }
}

module.exports = {
  writeToConsole,
  writeToFileAsync,
};
