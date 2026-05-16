function createSyntaxErrorMessage (line, column, msg) {
  return `Syntax error at line ${line} column ${column}. ${msg}`;
}

function formatParserV2ErrorMessage (error) {
  if (!error.diags) {
    // this is a runtime error
    return '';
  }

  const messageList = [];
  error.diags.forEach((diag) => {
    messageList.push(createSyntaxErrorMessage(diag.location.start.line, diag.location.start.column, diag.message));
  });

  return messageList.join('\n');
}

module.exports = {
  formatParserV2ErrorMessage,
};
