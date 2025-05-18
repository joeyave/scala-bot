import { logger } from "@/helpers/logger.ts";
import { KeySignature, Transposer } from "chord-transposer";

export function transposeAllText(
  htmlString: string,
  fromKey: string,
  toKey: string,
): string {
  // Parse the HTML into a DOM structure
  const parser = new DOMParser();
  const doc = parser.parseFromString(htmlString, "text/html");

  const toKeyParsed: KeySignature = Transposer.transpose(toKey).getKey();
  let fromKeyParsed: KeySignature;

  try {
    fromKeyParsed = Transposer.transpose(fromKey).getKey();
  } catch (err) {
    logger.error("Error parsing fromKey:", { err });
  }

  try {
    // Function to process all text nodes
    function processTextNodes(node: Node) {
      // If this is a text node with content
      if (
        node.nodeType === Node.TEXT_NODE &&
        node.textContent &&
        node.textContent.trim()
      ) {
        try {
          if (!fromKeyParsed) {
            try {
              fromKeyParsed = Transposer.transpose(node.textContent).getKey();
            } catch (err) {
              logger.error("Error parsing fromKey:", { err });
              return;
            }
          }

          // Try to transpose the text content
          // Replace the text content
          node.textContent = Transposer.transpose(node.textContent)
            .fromKey(fromKeyParsed)
            .toKey(toKeyParsed.majorKey)
            .toString();
        } catch (error) {
          // If this specific text node can't be transposed, leave it as is
          console.warn(`Couldn't transpose text: "${node.textContent}"`, error);
        }
      }

      // Recursively process all child nodes
      if (node.childNodes && node.childNodes.length > 0) {
        Array.from(node.childNodes).forEach(processTextNodes);
      }
    }

    // Start processing from the body
    processTextNodes(doc.body);

    // Return the modified HTML
    return doc.body.innerHTML;
  } catch (error) {
    // If overall transposition fails, throw the error
    console.error("Failed to transpose HTML text nodes", error);
    throw error;
  }
}
