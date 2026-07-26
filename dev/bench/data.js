window.BENCHMARK_DATA = {
  "lastUpdate": 1785049405988,
  "repoUrl": "https://github.com/alex60217101990/terse",
  "entries": {
    "terse Go Benchmarks": [
      {
        "commit": {
          "author": {
            "email": "33520849+alex60217101990@users.noreply.github.com",
            "name": "alex60217101990",
            "username": "alex60217101990"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "df94c081159ce44408205623c666d1f36618c82b",
          "message": "Merge pull request #19 from alex60217101990/ci/bench-trend-pages\n\nci: benchmark trend dashboard + regression alert (Pages)",
          "timestamp": "2026-07-26T10:01:42+03:00",
          "tree_id": "2e646f3e0956e22968cb4fb031ccabe093cdede7",
          "url": "https://github.com/alex60217101990/terse/commit/df94c081159ce44408205623c666d1f36618c82b"
        },
        "date": 1785049405310,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkRecord (github.com/alex60217101990/terse/internal/analytics)",
            "value": 10597,
            "unit": "ns/op\t     632 B/op\t       7 allocs/op",
            "extra": "214017 times\n4 procs"
          },
          {
            "name": "BenchmarkRecord (github.com/alex60217101990/terse/internal/analytics) - ns/op",
            "value": 10597,
            "unit": "ns/op",
            "extra": "214017 times\n4 procs"
          },
          {
            "name": "BenchmarkRecord (github.com/alex60217101990/terse/internal/analytics) - B/op",
            "value": 632,
            "unit": "B/op",
            "extra": "214017 times\n4 procs"
          },
          {
            "name": "BenchmarkRecord (github.com/alex60217101990/terse/internal/analytics) - allocs/op",
            "value": 7,
            "unit": "allocs/op",
            "extra": "214017 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff (github.com/alex60217101990/terse/internal/cache)",
            "value": 77182,
            "unit": "ns/op\t  426967 B/op\t      38 allocs/op",
            "extra": "30938 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 77182,
            "unit": "ns/op",
            "extra": "30938 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 426967,
            "unit": "B/op",
            "extra": "30938 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 38,
            "unit": "allocs/op",
            "extra": "30938 times\n4 procs"
          },
          {
            "name": "BenchmarkUtilityScore (github.com/alex60217101990/terse/internal/cache)",
            "value": 58.26,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "40881836 times\n4 procs"
          },
          {
            "name": "BenchmarkUtilityScore (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 58.26,
            "unit": "ns/op",
            "extra": "40881836 times\n4 procs"
          },
          {
            "name": "BenchmarkUtilityScore (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "40881836 times\n4 procs"
          },
          {
            "name": "BenchmarkUtilityScore (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "40881836 times\n4 procs"
          },
          {
            "name": "BenchmarkEvict_200Files (github.com/alex60217101990/terse/internal/cache)",
            "value": 35834,
            "unit": "ns/op\t    8193 B/op\t       1 allocs/op",
            "extra": "68834 times\n4 procs"
          },
          {
            "name": "BenchmarkEvict_200Files (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 35834,
            "unit": "ns/op",
            "extra": "68834 times\n4 procs"
          },
          {
            "name": "BenchmarkEvict_200Files (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 8193,
            "unit": "B/op",
            "extra": "68834 times\n4 procs"
          },
          {
            "name": "BenchmarkEvict_200Files (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "68834 times\n4 procs"
          },
          {
            "name": "BenchmarkDedup_Hit (github.com/alex60217101990/terse/internal/cache)",
            "value": 4387,
            "unit": "ns/op\t     728 B/op\t      10 allocs/op",
            "extra": "527781 times\n4 procs"
          },
          {
            "name": "BenchmarkDedup_Hit (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 4387,
            "unit": "ns/op",
            "extra": "527781 times\n4 procs"
          },
          {
            "name": "BenchmarkDedup_Hit (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 728,
            "unit": "B/op",
            "extra": "527781 times\n4 procs"
          },
          {
            "name": "BenchmarkDedup_Hit (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 10,
            "unit": "allocs/op",
            "extra": "527781 times\n4 procs"
          },
          {
            "name": "BenchmarkDedup_Miss (github.com/alex60217101990/terse/internal/cache)",
            "value": 40100,
            "unit": "ns/op\t    5133 B/op\t      16 allocs/op",
            "extra": "59936 times\n4 procs"
          },
          {
            "name": "BenchmarkDedup_Miss (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 40100,
            "unit": "ns/op",
            "extra": "59936 times\n4 procs"
          },
          {
            "name": "BenchmarkDedup_Miss (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 5133,
            "unit": "B/op",
            "extra": "59936 times\n4 procs"
          },
          {
            "name": "BenchmarkDedup_Miss (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 16,
            "unit": "allocs/op",
            "extra": "59936 times\n4 procs"
          },
          {
            "name": "BenchmarkRefGet (github.com/alex60217101990/terse/internal/cache)",
            "value": 9758,
            "unit": "ns/op\t    2133 B/op\t      11 allocs/op",
            "extra": "241114 times\n4 procs"
          },
          {
            "name": "BenchmarkRefGet (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 9758,
            "unit": "ns/op",
            "extra": "241114 times\n4 procs"
          },
          {
            "name": "BenchmarkRefGet (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 2133,
            "unit": "B/op",
            "extra": "241114 times\n4 procs"
          },
          {
            "name": "BenchmarkRefGet (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 11,
            "unit": "allocs/op",
            "extra": "241114 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveLoad (github.com/alex60217101990/terse/internal/cache)",
            "value": 225856,
            "unit": "ns/op\t   21865 B/op\t      25 allocs/op",
            "extra": "10000 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveLoad (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 225856,
            "unit": "ns/op",
            "extra": "10000 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveLoad (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 21865,
            "unit": "B/op",
            "extra": "10000 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveLoad (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 25,
            "unit": "allocs/op",
            "extra": "10000 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_ReadAll (github.com/alex60217101990/terse/internal/daemon)",
            "value": 4348,
            "unit": "ns/op\t   27824 B/op\t      12 allocs/op",
            "extra": "533887 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_ReadAll (github.com/alex60217101990/terse/internal/daemon) - ns/op",
            "value": 4348,
            "unit": "ns/op",
            "extra": "533887 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_ReadAll (github.com/alex60217101990/terse/internal/daemon) - B/op",
            "value": 27824,
            "unit": "B/op",
            "extra": "533887 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_ReadAll (github.com/alex60217101990/terse/internal/daemon) - allocs/op",
            "value": 12,
            "unit": "allocs/op",
            "extra": "533887 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_Pooled (github.com/alex60217101990/terse/internal/daemon)",
            "value": 185.2,
            "unit": "ns/op\t      48 B/op\t       1 allocs/op",
            "extra": "12862083 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_Pooled (github.com/alex60217101990/terse/internal/daemon) - ns/op",
            "value": 185.2,
            "unit": "ns/op",
            "extra": "12862083 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_Pooled (github.com/alex60217101990/terse/internal/daemon) - B/op",
            "value": 48,
            "unit": "B/op",
            "extra": "12862083 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_Pooled (github.com/alex60217101990/terse/internal/daemon) - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "12862083 times\n4 procs"
          },
          {
            "name": "BenchmarkDaemonRoundtrip (github.com/alex60217101990/terse/internal/daemon)",
            "value": 146624,
            "unit": "ns/op\t    7805 B/op\t      64 allocs/op",
            "extra": "16328 times\n4 procs"
          },
          {
            "name": "BenchmarkDaemonRoundtrip (github.com/alex60217101990/terse/internal/daemon) - ns/op",
            "value": 146624,
            "unit": "ns/op",
            "extra": "16328 times\n4 procs"
          },
          {
            "name": "BenchmarkDaemonRoundtrip (github.com/alex60217101990/terse/internal/daemon) - B/op",
            "value": 7805,
            "unit": "B/op",
            "extra": "16328 times\n4 procs"
          },
          {
            "name": "BenchmarkDaemonRoundtrip (github.com/alex60217101990/terse/internal/daemon) - allocs/op",
            "value": 64,
            "unit": "allocs/op",
            "extra": "16328 times\n4 procs"
          },
          {
            "name": "BenchmarkCLIRoundtrip (github.com/alex60217101990/terse/internal/daemon)",
            "value": 2829908,
            "unit": "ns/op\t    4161 B/op\t      55 allocs/op",
            "extra": "849 times\n4 procs"
          },
          {
            "name": "BenchmarkCLIRoundtrip (github.com/alex60217101990/terse/internal/daemon) - ns/op",
            "value": 2829908,
            "unit": "ns/op",
            "extra": "849 times\n4 procs"
          },
          {
            "name": "BenchmarkCLIRoundtrip (github.com/alex60217101990/terse/internal/daemon) - B/op",
            "value": 4161,
            "unit": "B/op",
            "extra": "849 times\n4 procs"
          },
          {
            "name": "BenchmarkCLIRoundtrip (github.com/alex60217101990/terse/internal/daemon) - allocs/op",
            "value": 55,
            "unit": "allocs/op",
            "extra": "849 times\n4 procs"
          },
          {
            "name": "BenchmarkAnalyzeJSONArray (github.com/alex60217101990/terse/internal/detect)",
            "value": 546166,
            "unit": "ns/op\t   18769 B/op\t      44 allocs/op",
            "extra": "4356 times\n4 procs"
          },
          {
            "name": "BenchmarkAnalyzeJSONArray (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 546166,
            "unit": "ns/op",
            "extra": "4356 times\n4 procs"
          },
          {
            "name": "BenchmarkAnalyzeJSONArray (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 18769,
            "unit": "B/op",
            "extra": "4356 times\n4 procs"
          },
          {
            "name": "BenchmarkAnalyzeJSONArray (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 44,
            "unit": "allocs/op",
            "extra": "4356 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Clean (github.com/alex60217101990/terse/internal/detect)",
            "value": 5759,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "415220 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Clean (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 5759,
            "unit": "ns/op",
            "extra": "415220 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Clean (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "415220 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Clean (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "415220 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Dirty (github.com/alex60217101990/terse/internal/detect)",
            "value": 5699,
            "unit": "ns/op\t    2688 B/op\t       1 allocs/op",
            "extra": "411387 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Dirty (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 5699,
            "unit": "ns/op",
            "extra": "411387 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Dirty (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 2688,
            "unit": "B/op",
            "extra": "411387 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Dirty (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "411387 times\n4 procs"
          },
          {
            "name": "BenchmarkSqueezeOutput (github.com/alex60217101990/terse/internal/detect)",
            "value": 150793,
            "unit": "ns/op\t   50764 B/op\t      18 allocs/op",
            "extra": "15828 times\n4 procs"
          },
          {
            "name": "BenchmarkSqueezeOutput (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 150793,
            "unit": "ns/op",
            "extra": "15828 times\n4 procs"
          },
          {
            "name": "BenchmarkSqueezeOutput (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 50764,
            "unit": "B/op",
            "extra": "15828 times\n4 procs"
          },
          {
            "name": "BenchmarkSqueezeOutput (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "15828 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_Unchanged (github.com/alex60217101990/terse/internal/hook)",
            "value": 320224,
            "unit": "ns/op\t   79705 B/op\t      76 allocs/op",
            "extra": "7586 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_Unchanged (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 320224,
            "unit": "ns/op",
            "extra": "7586 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_Unchanged (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 79705,
            "unit": "B/op",
            "extra": "7586 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_Unchanged (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 76,
            "unit": "allocs/op",
            "extra": "7586 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_FirstRead (github.com/alex60217101990/terse/internal/hook)",
            "value": 281646,
            "unit": "ns/op\t  100099 B/op\t      93 allocs/op",
            "extra": "8458 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_FirstRead (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 281646,
            "unit": "ns/op",
            "extra": "8458 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_FirstRead (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 100099,
            "unit": "B/op",
            "extra": "8458 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_FirstRead (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 93,
            "unit": "allocs/op",
            "extra": "8458 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Unchanged (github.com/alex60217101990/terse/internal/hook)",
            "value": 28542,
            "unit": "ns/op\t    4889 B/op\t      45 allocs/op",
            "extra": "82623 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Unchanged (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 28542,
            "unit": "ns/op",
            "extra": "82623 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Unchanged (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 4889,
            "unit": "B/op",
            "extra": "82623 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Unchanged (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 45,
            "unit": "allocs/op",
            "extra": "82623 times\n4 procs"
          },
          {
            "name": "BenchmarkBashHook_JSONArray (github.com/alex60217101990/terse/internal/hook)",
            "value": 3359745,
            "unit": "ns/op\t  955700 B/op\t     132 allocs/op",
            "extra": "712 times\n4 procs"
          },
          {
            "name": "BenchmarkBashHook_JSONArray (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 3359745,
            "unit": "ns/op",
            "extra": "712 times\n4 procs"
          },
          {
            "name": "BenchmarkBashHook_JSONArray (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 955700,
            "unit": "B/op",
            "extra": "712 times\n4 procs"
          },
          {
            "name": "BenchmarkBashHook_JSONArray (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 132,
            "unit": "allocs/op",
            "extra": "712 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Allow (github.com/alex60217101990/terse/internal/hook)",
            "value": 19648,
            "unit": "ns/op\t    2880 B/op\t      38 allocs/op",
            "extra": "121669 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Allow (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 19648,
            "unit": "ns/op",
            "extra": "121669 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Allow (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 2880,
            "unit": "B/op",
            "extra": "121669 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Allow (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 38,
            "unit": "allocs/op",
            "extra": "121669 times\n4 procs"
          },
          {
            "name": "BenchmarkGlob_Tree (github.com/alex60217101990/terse/internal/hook)",
            "value": 54471,
            "unit": "ns/op\t   13573 B/op\t     154 allocs/op",
            "extra": "43686 times\n4 procs"
          },
          {
            "name": "BenchmarkGlob_Tree (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 54471,
            "unit": "ns/op",
            "extra": "43686 times\n4 procs"
          },
          {
            "name": "BenchmarkGlob_Tree (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 13573,
            "unit": "B/op",
            "extra": "43686 times\n4 procs"
          },
          {
            "name": "BenchmarkGlob_Tree (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 154,
            "unit": "allocs/op",
            "extra": "43686 times\n4 procs"
          },
          {
            "name": "BenchmarkWrite_Compress (github.com/alex60217101990/terse/internal/hook)",
            "value": 26798,
            "unit": "ns/op\t    6152 B/op\t      39 allocs/op",
            "extra": "90157 times\n4 procs"
          },
          {
            "name": "BenchmarkWrite_Compress (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 26798,
            "unit": "ns/op",
            "extra": "90157 times\n4 procs"
          },
          {
            "name": "BenchmarkWrite_Compress (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 6152,
            "unit": "B/op",
            "extra": "90157 times\n4 procs"
          },
          {
            "name": "BenchmarkWrite_Compress (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 39,
            "unit": "allocs/op",
            "extra": "90157 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_PreToolIntercept (github.com/alex60217101990/terse/internal/hook)",
            "value": 28963,
            "unit": "ns/op\t    4945 B/op\t      45 allocs/op",
            "extra": "82275 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_PreToolIntercept (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 28963,
            "unit": "ns/op",
            "extra": "82275 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_PreToolIntercept (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 4945,
            "unit": "B/op",
            "extra": "82275 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_PreToolIntercept (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 45,
            "unit": "allocs/op",
            "extra": "82275 times\n4 procs"
          },
          {
            "name": "BenchmarkGrep_Grouped (github.com/alex60217101990/terse/internal/hook)",
            "value": 527342,
            "unit": "ns/op\t  192689 B/op\t     152 allocs/op",
            "extra": "4468 times\n4 procs"
          },
          {
            "name": "BenchmarkGrep_Grouped (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 527342,
            "unit": "ns/op",
            "extra": "4468 times\n4 procs"
          },
          {
            "name": "BenchmarkGrep_Grouped (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 192689,
            "unit": "B/op",
            "extra": "4468 times\n4 procs"
          },
          {
            "name": "BenchmarkGrep_Grouped (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 152,
            "unit": "allocs/op",
            "extra": "4468 times\n4 procs"
          }
        ]
      }
    ]
  }
}