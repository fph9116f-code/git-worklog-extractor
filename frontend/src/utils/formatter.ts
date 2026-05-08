export function formatFiles(files: string[]): string {
  return files.join('\n')
}

export function parseRepoNames(input: string): string[] {
  return input
    .split(/[\n,，]+/)
    .map((item) => item.trim())
    .filter(Boolean)
}
