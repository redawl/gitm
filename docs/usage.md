# Usage tips

## Using The Packet Filter

Filters take the form of "key:-value". Key is part of the packet to match, value is the value of that part
to match, and `-` indicates whether to do the inverse of the filter. 

Examples:

status:400   - Only displays http packets that have a response status code of 400.

hostname:-example.com - Only displays packets that were heading toward or coming from any host other than
example.com.