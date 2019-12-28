# Nimona IDL Codegen

```ndl
package conversation

import nimona.io/crypto crypto

stream mochi.io/conversation {
    signed root object Created {
        name string
    }
    signed object TopicUpdated {
        topic string
        dependsOn repeated relationship
    }
    signed object MessageAdded {
        body string
        dependsOn repeated relationship
    }
    signed object MessageRemoved {
        removes relationship
        dependsOn repeated relationship
    }
}
```
