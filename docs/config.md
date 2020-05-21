# Configuration

## Design

The config mechanism applies the following logic:

1. Configuration settings are first loaded from the Agent code, to provide a
   default value for each setting.
2. Then the loader looks for overrides in the environment, and overwrites the
   default values with them.
3. Then the Agent starts a remote config loading loop, which will loop on
   the following steps: 
4. Check if the API wrapping has entered termination and, if so, exit 
   configuration.
5. Load a set of settings values from Bearer.
6. Apply the remaining settings (concurrent-safe). 
7. If the previous step changed any setting, emit a configuration update event.
8. Sleep for a configured duration (this a setting)
9. Loop back to step 4.


## Dependencies

- A non-decorated transport, to perform config requests to Bearer, only if there
  is at least one non-overridden setting.
- A logger, to report errors.
- Default configuration, to obtain the config endpoint URL and transport settings.
- An event dispatcher, to emit config update events.
