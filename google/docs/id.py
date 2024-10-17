import re

def extract_file_id(ref):
    """
    Extracts and returns the Google file ID from a document ID or a link.

    Args:
        ref (str): A Google file ID or a link containing the ID.

    Returns:
        str: The extracted Google file ID.
    """
    # If the input is already a document ID, return it as is
    if re.match(r"^[a-zA-Z0-9_-]{33,}$", ref):
        return ref 
    
    # Regular expression to match Google Drive document links
    pattern = r"(?:https?://(?:drive|docs)\.google\.com/(?:file/d/|document/d/|open\?id=|uc\?id=))([a-zA-Z0-9_-]{33,})"

    # Try to extract the document ID from the link
    match = re.search(pattern, ref)
    if match:
        return match.group(1)
    
    # If the input doesn't match a known pattern, raise an error
    raise ValueError("Invalid Google document ID or link format")
