const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage();

  page.on('console', msg => {
    if (msg.type() === 'error') console.log('Browser Error:', msg.text());
    else console.log('Browser Log:', msg.text());
  });
  page.on('pageerror', err => console.log('Page Error:', err.message));

  await page.goto('http://localhost:3000');
  await page.waitForTimeout(2000);

  const root = await page.$('#root');
  if (root) {
    const html = await page.evaluate(el => el.innerHTML, root);
    console.log('Root content exactly:', html.substring(0, 200));
  }

  await browser.close();
})();
