// Type declarations for text file imports (handled by Wrangler bundler)
declare module '*.txt' {
  const content: string
  export default content
}
