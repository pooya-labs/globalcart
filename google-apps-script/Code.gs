/**
 * GlobalCart translation bridge.
 * Deploy this Apps Script as a Web App (execute as you; access: anyone).
 * Store a random token in Project Settings > Script Properties:
 *   GLOBALCART_TOKEN = <at least 24 random characters>
 */
function doPost(e) {
  try {
    var body = JSON.parse((e && e.postData && e.postData.contents) || "{}");
    var expected = PropertiesService.getScriptProperties().getProperty("GLOBALCART_TOKEN");
    if (!expected || !safeEqual(String(body.token || ""), expected)) {
      return jsonOutput({ ok: false, error: "Unauthorized" });
    }

    var fields = body.fields || {};
    ["title", "description", "category", "brand"].forEach(function (key) {
      if (typeof fields[key] !== "string" || fields[key].length > 12000) {
        throw new Error("Invalid field: " + key);
      }
    });

    // Brand names and identifiers are deliberately preserved, not machine-translated.
    var translated = {
      title: LanguageApp.translate(fields.title, "en", "fa"),
      description: LanguageApp.translate(fields.description, "en", "fa"),
      category: LanguageApp.translate(fields.category, "en", "fa"),
      brand: fields.brand
    };
    return jsonOutput({ ok: true, fields: translated });
  } catch (error) {
    return jsonOutput({ ok: false, error: "Translation failed" });
  }
}

function safeEqual(a, b) {
  if (a.length !== b.length) return false;
  var result = 0;
  for (var i = 0; i < a.length; i++) result |= a.charCodeAt(i) ^ b.charCodeAt(i);
  return result === 0;
}

function jsonOutput(value) {
  return ContentService.createTextOutput(JSON.stringify(value))
    .setMimeType(ContentService.MimeType.JSON);
}
