// Optional: import this in app.module.ts if you prefer module-based setup.
// Most teams just use instrumentation.ts (the file above).
import { Module } from '@nestjs/common';

@Module({})
export class OtelModule {
  // OTel SDK is initialized in instrumentation.ts before the module system loads.
  // This module is a no-op placeholder for teams who want a visible OTel entry point
  // in their module tree.
}
