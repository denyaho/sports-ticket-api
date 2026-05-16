const { Command } = require('@oclif/core');
const fs = require('fs');
const path = require('path');
const ora = require('ora');
const parse = require('../utils/parse');
const { formatParserV2ErrorMessage } = require('../utils/error-formatter');

class ValidateCommand extends Command {
  async run () {
    const spinner = ora({});
    let filepath = null;
    try {
      const { args } = await this.parse(ValidateCommand);
      filepath = args.filepath;
      let content = '';
      content = fs.readFileSync(path.resolve(process.cwd(), filepath), 'utf-8');

      spinner.text = 'Validating file content';
      spinner.start();

      await parse(content);
      spinner.succeed('Validating file content');
      spinner.succeed('Done. Parse succeeded without errors.');
    } catch (error) {
      const rawMessage = formatParserV2ErrorMessage(error);
      const message = rawMessage ? `You have syntax error(s) in ${path.basename(filepath)}\n${rawMessage}` : error.message;

      if (spinner.isSpinning) {
        spinner.fail(`Failed: ${message}`);
        return;
      }
      this.error(message);
    }
  }
}

ValidateCommand.description = 'validate docs content';

ValidateCommand.flags = {};

ValidateCommand.args = [
  { name: 'filepath', description: 'dbml file path' },
];

module.exports = ValidateCommand;
