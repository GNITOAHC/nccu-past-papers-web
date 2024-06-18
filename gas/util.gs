// Function to get sheet
function getSheet(sheetName) {
  var sheet = SpreadsheetApp.getActiveSpreadsheet().getSheetByName(sheetName);
  if (!sheet) {
    sheet = SpreadsheetApp.getActiveSpreadsheet().insertSheet(sheetName);
  }
  return sheet;
}

// Function to get column index by given column
function getColumnIndex(sheet, column) {
  var data = sheet.getDataRange().getValues();
  for (var i = 0; i < data.length; i++) {
    if (data[0][i] == column) return i;
  }
}

// Function to get row on sheet (which is added by 1)
function getRow(sheet, column, value) {
  var col = getColumnIndex(sheet, column);
  var data = sheet.getDataRange().getValues();
  for (var i = 1; i < data.length; i++) {
    if (data[i][col] == value) return i + 1;
  }
  return -1;
}