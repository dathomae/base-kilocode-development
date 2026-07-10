const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage();

  page.on('console', msg => console.log('Browser log:', msg.text()));
  page.on('pageerror', err => console.log('Browser err:', err.message));

  await page.goto('http://localhost:3001');
  await page.waitForTimeout(4000);

  const wrap = await page.$('.name-col-wrap');
  if (wrap) {
    const wrapHtml = await page.evaluate(el => el.outerHTML, wrap);
    console.log('Wrap HTML:', wrapHtml);
  } else {
    console.log('No name-col-wrap found on 3001');
  }

  await browser.close();
})();
