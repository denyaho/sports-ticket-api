const { parentPort, workerData } = require('worker_threads');
const { Parser } = require('@dbml/core');

function parse (content) {
  try {
    const parser = new Parser();
    const databaseObject = parser.parse(content, 'dbmlv2');
    const normalizedDatabase = databaseObject.normalize();
    const schemas = databaseObject.schemas.map((schema) => ({
      name: schema.name,
      tables: schema.tables.map((table) => table.name),
    }));

    return {
      name: databaseObject.name,
      description: databaseObject.note,
      schemas,
      normalizedDatabase,
    };
  } catch (err) {
    // Worker will cast custom error to `Error` type so we map DBML errors to JSON here to fix it.
    // https://developer.mozilla.org/en-US/docs/Web/API/Web_Workers_API/Structured_clone_algorithm#supported_types
    const diags = err.diags.map((d) => ({
      location: d.location,
      message: d.message,
    }));
    // eslint-disable-next-line no-throw-literal
    throw { diags };
  }
}

const { content } = workerData;
const result = parse(content);
parentPort.postMessage(result);
