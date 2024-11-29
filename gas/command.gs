// Function to add a new record to the sheet
function addRecord(sheetName, record) {
  var sheet = getSheet(sheetName);
  sheet.appendRow(record);
}

// Function to delete a record from the sheet
function deleteRecord(sheetName, columnName, rowValue) {
  var sheet = getSheet(sheetName);
  var rowNum = getRow(sheet, columnName, rowValue);
  sheet.deleteRow(rowNum);
}

// Function to dump all data
function getAll(sheetName) {
  var sheet = getSheet(sheetName);
  data = sheet.getDataRange().getValues();
  data.shift();
  return data;
}

// Function to search records in the sheet
function searchRecords(sheetName, searchColumn, searchValue) {
  var sheet = getSheet(sheetName);
  var data = sheet.getDataRange().getValues();
  var results = [];

  // Determine the column index based on the search column name
  var columnIndex = getColumnIndex(sheet, searchColumn);
  if (columnIndex == -1) {
    Logger.log('Sheetname: %s with searchColumn: %s not found', sheetName, searchColumn)
    return []
  }

  // Iterate through data to find matching records
  for (var i = 1; i < data.length; i++) { // Skip header row (i = 1)
    if (data[i][columnIndex] && data[i][columnIndex].toString().toLowerCase() === searchValue.toLowerCase()) {
      results.push(data[i]);
    }
  }
  return results;
}