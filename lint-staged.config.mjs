const q = (f) => (/\s/.test(f) ? `"${f}"` : f);

export default {
  'apps/web/**/*.{ts,tsx}': (files) => {
    if (files.length === 0) return [];
    const args = files.map(q).join(' ');
    return [
      `npm --prefix apps/web exec -- eslint --max-warnings=0 --fix --config apps/web/eslint.config.js ${args}`,
      `npx prettier --write ${args}`,
    ];
  },
  'apps/web/**/*.{css,json,md}': (files) => (files.length ? [`npx prettier --write ${files.join(' ')}`] : []),
  'apps/api/**/*.go': (files) => (files.length ? [`gofmt -w ${files.join(' ')}`] : []),
};
