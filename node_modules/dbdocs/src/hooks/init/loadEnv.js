/* eslint-disable no-restricted-syntax */
/* eslint-disable guard-for-in */
const netrc = require('netrc-parser').default;
const { HOST_CONFIG_STATUS } = require('../../utils/constants');

// eslint-disable-next-line func-names
module.exports = async function () {
  // eslint-disable-next-line global-require
  require('dotenv').config();

  // Check netrc config
  await netrc.load();

  const machines = Object.keys(netrc.machines).map((key) => ({ apiUrl: key, ...netrc.machines[key] }));

  const matchedMachine = machines.find((machine) => machine.apiUrl && machine.webUrl && machine.status === HOST_CONFIG_STATUS.active);

  if (!matchedMachine) return;

  // Override env
  process.env.DBDOCS_HOST = matchedMachine.webUrl;
  process.env.DBDOCS_API_HOST = matchedMachine.apiUrl;
};
