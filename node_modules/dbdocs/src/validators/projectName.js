const isValidName = (name, allowOrgPattern = false) => {
  const pattern = '([A-Za-z0-9_\\-@.\\s]+)';
  const testPattern = allowOrgPattern ? `^(${pattern}/)?${pattern}$` : `^${pattern}$`;
  return (new RegExp(testPattern)).test(name);
};

module.exports = {
  isValidName,
};
