const { Command, Flags } = require('@oclif/core');
const ora = require('ora');
const netrc = require('netrc-parser').default;
const { HOST_CONFIG_STATUS } = require('../utils/constants');
const { tryParseHttpUrl } = require('../validators/url');

class SetCommand extends Command {
  async run () {
    const spinner = ora({});
    try {
      const { flags } = await this.parse(SetCommand);

      const { webUrl, apiUrl } = flags;

      const parseAPIUrlResult = tryParseHttpUrl(apiUrl);
      if (!parseAPIUrlResult.isSuccess) {
        spinner.fail(`API URL parse error: ${parseAPIUrlResult.data}`);
        return;
      }

      const parseWebUrlResult = tryParseHttpUrl(webUrl);
      if (!parseWebUrlResult.isSuccess) {
        spinner.fail(`Web URL parse error: ${parseWebUrlResult.data}`);
        return;
      }

      // load netrc config
      await netrc.load();

      spinner.text = 'Saving your configurations';
      spinner.start();

      // make all hosts are inactive
      Object.keys(netrc.machines).forEach((host) => {
        if (netrc.machines[host].webUrl) {
          netrc.machines[host].status = HOST_CONFIG_STATUS.inactive;
        }
      });

      const parsedAPIUrl = parseAPIUrlResult.data;
      const parsedWebUrl = parseWebUrlResult.data;

      if (!netrc.machines[parsedAPIUrl]) {
        netrc.machines[parsedAPIUrl] = {};
      }

      netrc.machines[parsedAPIUrl].webUrl = parsedWebUrl;
      netrc.machines[parsedAPIUrl].status = HOST_CONFIG_STATUS.active;

      await netrc.save();

      spinner.succeed('Your configuration has been saved.');
    } catch (error) {
      if (spinner.isSpinning) {
        spinner.fail();
      }

      this.error(error.message);
    }
  }
}

SetCommand.description = 'Set the Web URL and API URL of dbdocs self-hosted server';

SetCommand.flags = {
  webUrl: Flags.string({
    description: 'Self-hosted web url',
    required: true,
  }),
  apiUrl: Flags.string({
    description: 'Self-hosted api url',
    required: true,
  }),
};

SetCommand.examples = [
  '$ dbdocs set --webUrl http://webserver.dev --apiUrl http://apiserver.dev',
].join('\n');

module.exports = SetCommand;
