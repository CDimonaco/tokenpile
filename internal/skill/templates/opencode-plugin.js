// tokenpile capture plugin for opencode.
//
// Subscribes to session.idle and forwards the usage opencode recorded for that
// session to `tokenpile hook opencode`. The numbers come from opencode's own
// record of what the provider reported: nothing here is estimated.
//
// Installed and removed by `tokenpile skill install|uninstall opencode`.
export const tokenpile = async ({ client, directory, $ }) => {
  return {
    event: async ({ event }) => {
      if (event.type !== "session.idle") return;

      const sessionID = event.properties?.sessionID;
      if (!sessionID) return;

      let messages = [];
      try {
        const res = await client.session.messages({ path: { id: sessionID } });
        messages = res?.data ?? res ?? [];
      } catch (err) {
        console.error("tokenpile: could not read session messages", err);
        return;
      }

      let branch = "";
      try {
        branch = (await $`git -C ${directory} rev-parse --abbrev-ref HEAD`.text()).trim();
      } catch {
        // Not a git repository, or no HEAD yet. Attribution falls back to a
        // binding, and failing that the usage is recorded unattributed.
      }

      const payload = JSON.stringify({
        session_id: sessionID,
        branch,
        messages: messages.map((m) => ({
          id: m.id,
          role: m.role ?? m.info?.role,
          modelID: m.modelID ?? m.info?.modelID,
          path: m.path ?? m.info?.path ?? { cwd: directory },
          tokens: m.tokens ?? m.info?.tokens,
          time: m.time ?? m.info?.time,
        })),
      });

      try {
        await $`tokenpile hook opencode`.stdin(payload).quiet();
      } catch (err) {
        console.error("tokenpile: capture failed", err);
      }
    },
  };
};
