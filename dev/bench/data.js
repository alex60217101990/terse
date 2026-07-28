window.BENCHMARK_DATA = {
  "lastUpdate": 1785267671330,
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
      },
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
          "id": "1fab20d5f755397367804e9bc74f6ffc31d4ed01",
          "message": "Merge pull request #22 from alex60217101990/feat/block-dedup\n\nfeat(detect): fold non-adjacent duplicate blocks in generic output",
          "timestamp": "2026-07-27T14:41:08+03:00",
          "tree_id": "af9b0df3f096b1780fd9f7773a920b63427e33de",
          "url": "https://github.com/alex60217101990/terse/commit/1fab20d5f755397367804e9bc74f6ffc31d4ed01"
        },
        "date": 1785152561132,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkRecord (github.com/alex60217101990/terse/internal/analytics)",
            "value": 10040,
            "unit": "ns/op\t     632 B/op\t       7 allocs/op",
            "extra": "238462 times\n4 procs"
          },
          {
            "name": "BenchmarkRecord (github.com/alex60217101990/terse/internal/analytics) - ns/op",
            "value": 10040,
            "unit": "ns/op",
            "extra": "238462 times\n4 procs"
          },
          {
            "name": "BenchmarkRecord (github.com/alex60217101990/terse/internal/analytics) - B/op",
            "value": 632,
            "unit": "B/op",
            "extra": "238462 times\n4 procs"
          },
          {
            "name": "BenchmarkRecord (github.com/alex60217101990/terse/internal/analytics) - allocs/op",
            "value": 7,
            "unit": "allocs/op",
            "extra": "238462 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff (github.com/alex60217101990/terse/internal/cache)",
            "value": 64097,
            "unit": "ns/op\t  426954 B/op\t      38 allocs/op",
            "extra": "37044 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 64097,
            "unit": "ns/op",
            "extra": "37044 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 426954,
            "unit": "B/op",
            "extra": "37044 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 38,
            "unit": "allocs/op",
            "extra": "37044 times\n4 procs"
          },
          {
            "name": "BenchmarkUtilityScore (github.com/alex60217101990/terse/internal/cache)",
            "value": 47.45,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "51802324 times\n4 procs"
          },
          {
            "name": "BenchmarkUtilityScore (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 47.45,
            "unit": "ns/op",
            "extra": "51802324 times\n4 procs"
          },
          {
            "name": "BenchmarkUtilityScore (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "51802324 times\n4 procs"
          },
          {
            "name": "BenchmarkUtilityScore (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "51802324 times\n4 procs"
          },
          {
            "name": "BenchmarkEvict_200Files (github.com/alex60217101990/terse/internal/cache)",
            "value": 26572,
            "unit": "ns/op\t    8193 B/op\t       1 allocs/op",
            "extra": "89319 times\n4 procs"
          },
          {
            "name": "BenchmarkEvict_200Files (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 26572,
            "unit": "ns/op",
            "extra": "89319 times\n4 procs"
          },
          {
            "name": "BenchmarkEvict_200Files (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 8193,
            "unit": "B/op",
            "extra": "89319 times\n4 procs"
          },
          {
            "name": "BenchmarkEvict_200Files (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "89319 times\n4 procs"
          },
          {
            "name": "BenchmarkDedup_Hit (github.com/alex60217101990/terse/internal/cache)",
            "value": 3875,
            "unit": "ns/op\t     728 B/op\t      10 allocs/op",
            "extra": "614154 times\n4 procs"
          },
          {
            "name": "BenchmarkDedup_Hit (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 3875,
            "unit": "ns/op",
            "extra": "614154 times\n4 procs"
          },
          {
            "name": "BenchmarkDedup_Hit (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 728,
            "unit": "B/op",
            "extra": "614154 times\n4 procs"
          },
          {
            "name": "BenchmarkDedup_Hit (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 10,
            "unit": "allocs/op",
            "extra": "614154 times\n4 procs"
          },
          {
            "name": "BenchmarkDedup_Miss (github.com/alex60217101990/terse/internal/cache)",
            "value": 31895,
            "unit": "ns/op\t    5134 B/op\t      16 allocs/op",
            "extra": "78781 times\n4 procs"
          },
          {
            "name": "BenchmarkDedup_Miss (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 31895,
            "unit": "ns/op",
            "extra": "78781 times\n4 procs"
          },
          {
            "name": "BenchmarkDedup_Miss (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 5134,
            "unit": "B/op",
            "extra": "78781 times\n4 procs"
          },
          {
            "name": "BenchmarkDedup_Miss (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 16,
            "unit": "allocs/op",
            "extra": "78781 times\n4 procs"
          },
          {
            "name": "BenchmarkRefGet (github.com/alex60217101990/terse/internal/cache)",
            "value": 9782,
            "unit": "ns/op\t    2132 B/op\t      11 allocs/op",
            "extra": "238296 times\n4 procs"
          },
          {
            "name": "BenchmarkRefGet (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 9782,
            "unit": "ns/op",
            "extra": "238296 times\n4 procs"
          },
          {
            "name": "BenchmarkRefGet (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 2132,
            "unit": "B/op",
            "extra": "238296 times\n4 procs"
          },
          {
            "name": "BenchmarkRefGet (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 11,
            "unit": "allocs/op",
            "extra": "238296 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveLoad (github.com/alex60217101990/terse/internal/cache)",
            "value": 170295,
            "unit": "ns/op\t   21875 B/op\t      25 allocs/op",
            "extra": "14874 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveLoad (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 170295,
            "unit": "ns/op",
            "extra": "14874 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveLoad (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 21875,
            "unit": "B/op",
            "extra": "14874 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveLoad (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 25,
            "unit": "allocs/op",
            "extra": "14874 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_ReadAll (github.com/alex60217101990/terse/internal/daemon)",
            "value": 3505,
            "unit": "ns/op\t   27824 B/op\t      12 allocs/op",
            "extra": "671050 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_ReadAll (github.com/alex60217101990/terse/internal/daemon) - ns/op",
            "value": 3505,
            "unit": "ns/op",
            "extra": "671050 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_ReadAll (github.com/alex60217101990/terse/internal/daemon) - B/op",
            "value": 27824,
            "unit": "B/op",
            "extra": "671050 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_ReadAll (github.com/alex60217101990/terse/internal/daemon) - allocs/op",
            "value": 12,
            "unit": "allocs/op",
            "extra": "671050 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_Pooled (github.com/alex60217101990/terse/internal/daemon)",
            "value": 158.3,
            "unit": "ns/op\t      48 B/op\t       1 allocs/op",
            "extra": "15177494 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_Pooled (github.com/alex60217101990/terse/internal/daemon) - ns/op",
            "value": 158.3,
            "unit": "ns/op",
            "extra": "15177494 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_Pooled (github.com/alex60217101990/terse/internal/daemon) - B/op",
            "value": 48,
            "unit": "B/op",
            "extra": "15177494 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_Pooled (github.com/alex60217101990/terse/internal/daemon) - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "15177494 times\n4 procs"
          },
          {
            "name": "BenchmarkDaemonRoundtrip (github.com/alex60217101990/terse/internal/daemon)",
            "value": 113817,
            "unit": "ns/op\t    7800 B/op\t      64 allocs/op",
            "extra": "20438 times\n4 procs"
          },
          {
            "name": "BenchmarkDaemonRoundtrip (github.com/alex60217101990/terse/internal/daemon) - ns/op",
            "value": 113817,
            "unit": "ns/op",
            "extra": "20438 times\n4 procs"
          },
          {
            "name": "BenchmarkDaemonRoundtrip (github.com/alex60217101990/terse/internal/daemon) - B/op",
            "value": 7800,
            "unit": "B/op",
            "extra": "20438 times\n4 procs"
          },
          {
            "name": "BenchmarkDaemonRoundtrip (github.com/alex60217101990/terse/internal/daemon) - allocs/op",
            "value": 64,
            "unit": "allocs/op",
            "extra": "20438 times\n4 procs"
          },
          {
            "name": "BenchmarkCLIRoundtrip (github.com/alex60217101990/terse/internal/daemon)",
            "value": 2552696,
            "unit": "ns/op\t    4148 B/op\t      55 allocs/op",
            "extra": "943 times\n4 procs"
          },
          {
            "name": "BenchmarkCLIRoundtrip (github.com/alex60217101990/terse/internal/daemon) - ns/op",
            "value": 2552696,
            "unit": "ns/op",
            "extra": "943 times\n4 procs"
          },
          {
            "name": "BenchmarkCLIRoundtrip (github.com/alex60217101990/terse/internal/daemon) - B/op",
            "value": 4148,
            "unit": "B/op",
            "extra": "943 times\n4 procs"
          },
          {
            "name": "BenchmarkCLIRoundtrip (github.com/alex60217101990/terse/internal/daemon) - allocs/op",
            "value": 55,
            "unit": "allocs/op",
            "extra": "943 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_NoDup (github.com/alex60217101990/terse/internal/detect)",
            "value": 1095,
            "unit": "ns/op\t 745.52 MB/s\t     456 B/op\t       3 allocs/op",
            "extra": "2192725 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_NoDup (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 1095,
            "unit": "ns/op",
            "extra": "2192725 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_NoDup (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 745.52,
            "unit": "MB/s",
            "extra": "2192725 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_NoDup (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 456,
            "unit": "B/op",
            "extra": "2192725 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_NoDup (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 3,
            "unit": "allocs/op",
            "extra": "2192725 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks (github.com/alex60217101990/terse/internal/detect)",
            "value": 1075,
            "unit": "ns/op\t1067.76 MB/s\t    1152 B/op\t       1 allocs/op",
            "extra": "2222421 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 1075,
            "unit": "ns/op",
            "extra": "2222421 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 1067.76,
            "unit": "MB/s",
            "extra": "2222421 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 1152,
            "unit": "B/op",
            "extra": "2222421 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "2222421 times\n4 procs"
          },
          {
            "name": "BenchmarkAnalyzeJSONArray (github.com/alex60217101990/terse/internal/detect)",
            "value": 407734,
            "unit": "ns/op\t   18768 B/op\t      44 allocs/op",
            "extra": "5900 times\n4 procs"
          },
          {
            "name": "BenchmarkAnalyzeJSONArray (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 407734,
            "unit": "ns/op",
            "extra": "5900 times\n4 procs"
          },
          {
            "name": "BenchmarkAnalyzeJSONArray (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 18768,
            "unit": "B/op",
            "extra": "5900 times\n4 procs"
          },
          {
            "name": "BenchmarkAnalyzeJSONArray (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 44,
            "unit": "allocs/op",
            "extra": "5900 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Clean (github.com/alex60217101990/terse/internal/detect)",
            "value": 4671,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "512559 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Clean (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 4671,
            "unit": "ns/op",
            "extra": "512559 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Clean (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "512559 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Clean (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "512559 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Dirty (github.com/alex60217101990/terse/internal/detect)",
            "value": 4441,
            "unit": "ns/op\t    2688 B/op\t       1 allocs/op",
            "extra": "519154 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Dirty (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 4441,
            "unit": "ns/op",
            "extra": "519154 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Dirty (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 2688,
            "unit": "B/op",
            "extra": "519154 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Dirty (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "519154 times\n4 procs"
          },
          {
            "name": "BenchmarkSqueezeOutput (github.com/alex60217101990/terse/internal/detect)",
            "value": 115448,
            "unit": "ns/op\t   50798 B/op\t      18 allocs/op",
            "extra": "20830 times\n4 procs"
          },
          {
            "name": "BenchmarkSqueezeOutput (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 115448,
            "unit": "ns/op",
            "extra": "20830 times\n4 procs"
          },
          {
            "name": "BenchmarkSqueezeOutput (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 50798,
            "unit": "B/op",
            "extra": "20830 times\n4 procs"
          },
          {
            "name": "BenchmarkSqueezeOutput (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "20830 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_Unchanged (github.com/alex60217101990/terse/internal/hook)",
            "value": 255731,
            "unit": "ns/op\t   79667 B/op\t      76 allocs/op",
            "extra": "10000 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_Unchanged (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 255731,
            "unit": "ns/op",
            "extra": "10000 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_Unchanged (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 79667,
            "unit": "B/op",
            "extra": "10000 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_Unchanged (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 76,
            "unit": "allocs/op",
            "extra": "10000 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_FirstRead (github.com/alex60217101990/terse/internal/hook)",
            "value": 213890,
            "unit": "ns/op\t  100152 B/op\t      93 allocs/op",
            "extra": "10000 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_FirstRead (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 213890,
            "unit": "ns/op",
            "extra": "10000 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_FirstRead (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 100152,
            "unit": "B/op",
            "extra": "10000 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_FirstRead (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 93,
            "unit": "allocs/op",
            "extra": "10000 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Unchanged (github.com/alex60217101990/terse/internal/hook)",
            "value": 25317,
            "unit": "ns/op\t    4889 B/op\t      45 allocs/op",
            "extra": "93259 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Unchanged (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 25317,
            "unit": "ns/op",
            "extra": "93259 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Unchanged (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 4889,
            "unit": "B/op",
            "extra": "93259 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Unchanged (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 45,
            "unit": "allocs/op",
            "extra": "93259 times\n4 procs"
          },
          {
            "name": "BenchmarkBashHook_JSONArray (github.com/alex60217101990/terse/internal/hook)",
            "value": 2628999,
            "unit": "ns/op\t  955587 B/op\t     131 allocs/op",
            "extra": "914 times\n4 procs"
          },
          {
            "name": "BenchmarkBashHook_JSONArray (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 2628999,
            "unit": "ns/op",
            "extra": "914 times\n4 procs"
          },
          {
            "name": "BenchmarkBashHook_JSONArray (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 955587,
            "unit": "B/op",
            "extra": "914 times\n4 procs"
          },
          {
            "name": "BenchmarkBashHook_JSONArray (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 131,
            "unit": "allocs/op",
            "extra": "914 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Allow (github.com/alex60217101990/terse/internal/hook)",
            "value": 17423,
            "unit": "ns/op\t    2880 B/op\t      38 allocs/op",
            "extra": "137775 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Allow (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 17423,
            "unit": "ns/op",
            "extra": "137775 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Allow (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 2880,
            "unit": "B/op",
            "extra": "137775 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Allow (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 38,
            "unit": "allocs/op",
            "extra": "137775 times\n4 procs"
          },
          {
            "name": "BenchmarkGlob_Tree (github.com/alex60217101990/terse/internal/hook)",
            "value": 42430,
            "unit": "ns/op\t   13574 B/op\t     154 allocs/op",
            "extra": "56769 times\n4 procs"
          },
          {
            "name": "BenchmarkGlob_Tree (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 42430,
            "unit": "ns/op",
            "extra": "56769 times\n4 procs"
          },
          {
            "name": "BenchmarkGlob_Tree (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 13574,
            "unit": "B/op",
            "extra": "56769 times\n4 procs"
          },
          {
            "name": "BenchmarkGlob_Tree (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 154,
            "unit": "allocs/op",
            "extra": "56769 times\n4 procs"
          },
          {
            "name": "BenchmarkWrite_Compress (github.com/alex60217101990/terse/internal/hook)",
            "value": 22359,
            "unit": "ns/op\t    6152 B/op\t      39 allocs/op",
            "extra": "107991 times\n4 procs"
          },
          {
            "name": "BenchmarkWrite_Compress (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 22359,
            "unit": "ns/op",
            "extra": "107991 times\n4 procs"
          },
          {
            "name": "BenchmarkWrite_Compress (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 6152,
            "unit": "B/op",
            "extra": "107991 times\n4 procs"
          },
          {
            "name": "BenchmarkWrite_Compress (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 39,
            "unit": "allocs/op",
            "extra": "107991 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_PreToolIntercept (github.com/alex60217101990/terse/internal/hook)",
            "value": 25460,
            "unit": "ns/op\t    4897 B/op\t      45 allocs/op",
            "extra": "93721 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_PreToolIntercept (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 25460,
            "unit": "ns/op",
            "extra": "93721 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_PreToolIntercept (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 4897,
            "unit": "B/op",
            "extra": "93721 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_PreToolIntercept (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 45,
            "unit": "allocs/op",
            "extra": "93721 times\n4 procs"
          },
          {
            "name": "BenchmarkGrep_Grouped (github.com/alex60217101990/terse/internal/hook)",
            "value": 413472,
            "unit": "ns/op\t  192586 B/op\t     152 allocs/op",
            "extra": "5754 times\n4 procs"
          },
          {
            "name": "BenchmarkGrep_Grouped (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 413472,
            "unit": "ns/op",
            "extra": "5754 times\n4 procs"
          },
          {
            "name": "BenchmarkGrep_Grouped (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 192586,
            "unit": "B/op",
            "extra": "5754 times\n4 procs"
          },
          {
            "name": "BenchmarkGrep_Grouped (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 152,
            "unit": "allocs/op",
            "extra": "5754 times\n4 procs"
          }
        ]
      },
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
          "id": "d59578e99013f19433c2bf31def77c637cc7d991",
          "message": "Merge pull request #24 from alex60217101990/feat/compression-wave2\n\nfeat: compression wave 2 — 9 new detectors, recoverable summaries, max-perf hot paths",
          "timestamp": "2026-07-28T22:34:53+03:00",
          "tree_id": "3119dda5d26e3f8831a20e9f309bd07b26ebae49",
          "url": "https://github.com/alex60217101990/terse/commit/d59578e99013f19433c2bf31def77c637cc7d991"
        },
        "date": 1785267671034,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkRecord (github.com/alex60217101990/terse/internal/analytics)",
            "value": 13238,
            "unit": "ns/op\t     632 B/op\t       7 allocs/op",
            "extra": "90151 times\n4 procs"
          },
          {
            "name": "BenchmarkRecord (github.com/alex60217101990/terse/internal/analytics) - ns/op",
            "value": 13238,
            "unit": "ns/op",
            "extra": "90151 times\n4 procs"
          },
          {
            "name": "BenchmarkRecord (github.com/alex60217101990/terse/internal/analytics) - B/op",
            "value": 632,
            "unit": "B/op",
            "extra": "90151 times\n4 procs"
          },
          {
            "name": "BenchmarkRecord (github.com/alex60217101990/terse/internal/analytics) - allocs/op",
            "value": 7,
            "unit": "allocs/op",
            "extra": "90151 times\n4 procs"
          },
          {
            "name": "BenchmarkRecord (github.com/alex60217101990/terse/internal/analytics)",
            "value": 13062,
            "unit": "ns/op\t     632 B/op\t       7 allocs/op",
            "extra": "91690 times\n4 procs"
          },
          {
            "name": "BenchmarkRecord (github.com/alex60217101990/terse/internal/analytics) - ns/op",
            "value": 13062,
            "unit": "ns/op",
            "extra": "91690 times\n4 procs"
          },
          {
            "name": "BenchmarkRecord (github.com/alex60217101990/terse/internal/analytics) - B/op",
            "value": 632,
            "unit": "B/op",
            "extra": "91690 times\n4 procs"
          },
          {
            "name": "BenchmarkRecord (github.com/alex60217101990/terse/internal/analytics) - allocs/op",
            "value": 7,
            "unit": "allocs/op",
            "extra": "91690 times\n4 procs"
          },
          {
            "name": "BenchmarkRecord (github.com/alex60217101990/terse/internal/analytics)",
            "value": 12900,
            "unit": "ns/op\t     632 B/op\t       7 allocs/op",
            "extra": "92458 times\n4 procs"
          },
          {
            "name": "BenchmarkRecord (github.com/alex60217101990/terse/internal/analytics) - ns/op",
            "value": 12900,
            "unit": "ns/op",
            "extra": "92458 times\n4 procs"
          },
          {
            "name": "BenchmarkRecord (github.com/alex60217101990/terse/internal/analytics) - B/op",
            "value": 632,
            "unit": "B/op",
            "extra": "92458 times\n4 procs"
          },
          {
            "name": "BenchmarkRecord (github.com/alex60217101990/terse/internal/analytics) - allocs/op",
            "value": 7,
            "unit": "allocs/op",
            "extra": "92458 times\n4 procs"
          },
          {
            "name": "BenchmarkRecord (github.com/alex60217101990/terse/internal/analytics)",
            "value": 13102,
            "unit": "ns/op\t     632 B/op\t       7 allocs/op",
            "extra": "91921 times\n4 procs"
          },
          {
            "name": "BenchmarkRecord (github.com/alex60217101990/terse/internal/analytics) - ns/op",
            "value": 13102,
            "unit": "ns/op",
            "extra": "91921 times\n4 procs"
          },
          {
            "name": "BenchmarkRecord (github.com/alex60217101990/terse/internal/analytics) - B/op",
            "value": 632,
            "unit": "B/op",
            "extra": "91921 times\n4 procs"
          },
          {
            "name": "BenchmarkRecord (github.com/alex60217101990/terse/internal/analytics) - allocs/op",
            "value": 7,
            "unit": "allocs/op",
            "extra": "91921 times\n4 procs"
          },
          {
            "name": "BenchmarkRecord (github.com/alex60217101990/terse/internal/analytics)",
            "value": 12939,
            "unit": "ns/op\t     632 B/op\t       7 allocs/op",
            "extra": "91712 times\n4 procs"
          },
          {
            "name": "BenchmarkRecord (github.com/alex60217101990/terse/internal/analytics) - ns/op",
            "value": 12939,
            "unit": "ns/op",
            "extra": "91712 times\n4 procs"
          },
          {
            "name": "BenchmarkRecord (github.com/alex60217101990/terse/internal/analytics) - B/op",
            "value": 632,
            "unit": "B/op",
            "extra": "91712 times\n4 procs"
          },
          {
            "name": "BenchmarkRecord (github.com/alex60217101990/terse/internal/analytics) - allocs/op",
            "value": 7,
            "unit": "allocs/op",
            "extra": "91712 times\n4 procs"
          },
          {
            "name": "BenchmarkRecord (github.com/alex60217101990/terse/internal/analytics)",
            "value": 13022,
            "unit": "ns/op\t     632 B/op\t       7 allocs/op",
            "extra": "93020 times\n4 procs"
          },
          {
            "name": "BenchmarkRecord (github.com/alex60217101990/terse/internal/analytics) - ns/op",
            "value": 13022,
            "unit": "ns/op",
            "extra": "93020 times\n4 procs"
          },
          {
            "name": "BenchmarkRecord (github.com/alex60217101990/terse/internal/analytics) - B/op",
            "value": 632,
            "unit": "B/op",
            "extra": "93020 times\n4 procs"
          },
          {
            "name": "BenchmarkRecord (github.com/alex60217101990/terse/internal/analytics) - allocs/op",
            "value": 7,
            "unit": "allocs/op",
            "extra": "93020 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff (github.com/alex60217101990/terse/internal/cache)",
            "value": 24964,
            "unit": "ns/op\t   25364 B/op\t      13 allocs/op",
            "extra": "46448 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 24964,
            "unit": "ns/op",
            "extra": "46448 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 25364,
            "unit": "B/op",
            "extra": "46448 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 13,
            "unit": "allocs/op",
            "extra": "46448 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff (github.com/alex60217101990/terse/internal/cache)",
            "value": 26041,
            "unit": "ns/op\t   25342 B/op\t      13 allocs/op",
            "extra": "43617 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 26041,
            "unit": "ns/op",
            "extra": "43617 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 25342,
            "unit": "B/op",
            "extra": "43617 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 13,
            "unit": "allocs/op",
            "extra": "43617 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff (github.com/alex60217101990/terse/internal/cache)",
            "value": 24506,
            "unit": "ns/op\t   25341 B/op\t      13 allocs/op",
            "extra": "48523 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 24506,
            "unit": "ns/op",
            "extra": "48523 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 25341,
            "unit": "B/op",
            "extra": "48523 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 13,
            "unit": "allocs/op",
            "extra": "48523 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff (github.com/alex60217101990/terse/internal/cache)",
            "value": 24433,
            "unit": "ns/op\t   25341 B/op\t      13 allocs/op",
            "extra": "48734 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 24433,
            "unit": "ns/op",
            "extra": "48734 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 25341,
            "unit": "B/op",
            "extra": "48734 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 13,
            "unit": "allocs/op",
            "extra": "48734 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff (github.com/alex60217101990/terse/internal/cache)",
            "value": 24689,
            "unit": "ns/op\t   25341 B/op\t      13 allocs/op",
            "extra": "49281 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 24689,
            "unit": "ns/op",
            "extra": "49281 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 25341,
            "unit": "B/op",
            "extra": "49281 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 13,
            "unit": "allocs/op",
            "extra": "49281 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff (github.com/alex60217101990/terse/internal/cache)",
            "value": 25121,
            "unit": "ns/op\t   25341 B/op\t      13 allocs/op",
            "extra": "48699 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 25121,
            "unit": "ns/op",
            "extra": "48699 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 25341,
            "unit": "B/op",
            "extra": "48699 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 13,
            "unit": "allocs/op",
            "extra": "48699 times\n4 procs"
          },
          {
            "name": "BenchmarkUtilityScore (github.com/alex60217101990/terse/internal/cache)",
            "value": 59.39,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "20007325 times\n4 procs"
          },
          {
            "name": "BenchmarkUtilityScore (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 59.39,
            "unit": "ns/op",
            "extra": "20007325 times\n4 procs"
          },
          {
            "name": "BenchmarkUtilityScore (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "20007325 times\n4 procs"
          },
          {
            "name": "BenchmarkUtilityScore (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "20007325 times\n4 procs"
          },
          {
            "name": "BenchmarkUtilityScore (github.com/alex60217101990/terse/internal/cache)",
            "value": 59.36,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "20191744 times\n4 procs"
          },
          {
            "name": "BenchmarkUtilityScore (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 59.36,
            "unit": "ns/op",
            "extra": "20191744 times\n4 procs"
          },
          {
            "name": "BenchmarkUtilityScore (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "20191744 times\n4 procs"
          },
          {
            "name": "BenchmarkUtilityScore (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "20191744 times\n4 procs"
          },
          {
            "name": "BenchmarkUtilityScore (github.com/alex60217101990/terse/internal/cache)",
            "value": 59.34,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "20185546 times\n4 procs"
          },
          {
            "name": "BenchmarkUtilityScore (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 59.34,
            "unit": "ns/op",
            "extra": "20185546 times\n4 procs"
          },
          {
            "name": "BenchmarkUtilityScore (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "20185546 times\n4 procs"
          },
          {
            "name": "BenchmarkUtilityScore (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "20185546 times\n4 procs"
          },
          {
            "name": "BenchmarkUtilityScore (github.com/alex60217101990/terse/internal/cache)",
            "value": 59.52,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "20172120 times\n4 procs"
          },
          {
            "name": "BenchmarkUtilityScore (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 59.52,
            "unit": "ns/op",
            "extra": "20172120 times\n4 procs"
          },
          {
            "name": "BenchmarkUtilityScore (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "20172120 times\n4 procs"
          },
          {
            "name": "BenchmarkUtilityScore (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "20172120 times\n4 procs"
          },
          {
            "name": "BenchmarkUtilityScore (github.com/alex60217101990/terse/internal/cache)",
            "value": 59.35,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "20229706 times\n4 procs"
          },
          {
            "name": "BenchmarkUtilityScore (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 59.35,
            "unit": "ns/op",
            "extra": "20229706 times\n4 procs"
          },
          {
            "name": "BenchmarkUtilityScore (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "20229706 times\n4 procs"
          },
          {
            "name": "BenchmarkUtilityScore (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "20229706 times\n4 procs"
          },
          {
            "name": "BenchmarkUtilityScore (github.com/alex60217101990/terse/internal/cache)",
            "value": 59.46,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "20161214 times\n4 procs"
          },
          {
            "name": "BenchmarkUtilityScore (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 59.46,
            "unit": "ns/op",
            "extra": "20161214 times\n4 procs"
          },
          {
            "name": "BenchmarkUtilityScore (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "20161214 times\n4 procs"
          },
          {
            "name": "BenchmarkUtilityScore (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "20161214 times\n4 procs"
          },
          {
            "name": "BenchmarkEvict_200Files (github.com/alex60217101990/terse/internal/cache)",
            "value": 36348,
            "unit": "ns/op\t    8194 B/op\t       1 allocs/op",
            "extra": "32695 times\n4 procs"
          },
          {
            "name": "BenchmarkEvict_200Files (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 36348,
            "unit": "ns/op",
            "extra": "32695 times\n4 procs"
          },
          {
            "name": "BenchmarkEvict_200Files (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 8194,
            "unit": "B/op",
            "extra": "32695 times\n4 procs"
          },
          {
            "name": "BenchmarkEvict_200Files (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "32695 times\n4 procs"
          },
          {
            "name": "BenchmarkEvict_200Files (github.com/alex60217101990/terse/internal/cache)",
            "value": 35576,
            "unit": "ns/op\t    8194 B/op\t       1 allocs/op",
            "extra": "34131 times\n4 procs"
          },
          {
            "name": "BenchmarkEvict_200Files (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 35576,
            "unit": "ns/op",
            "extra": "34131 times\n4 procs"
          },
          {
            "name": "BenchmarkEvict_200Files (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 8194,
            "unit": "B/op",
            "extra": "34131 times\n4 procs"
          },
          {
            "name": "BenchmarkEvict_200Files (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "34131 times\n4 procs"
          },
          {
            "name": "BenchmarkEvict_200Files (github.com/alex60217101990/terse/internal/cache)",
            "value": 35242,
            "unit": "ns/op\t    8194 B/op\t       1 allocs/op",
            "extra": "34266 times\n4 procs"
          },
          {
            "name": "BenchmarkEvict_200Files (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 35242,
            "unit": "ns/op",
            "extra": "34266 times\n4 procs"
          },
          {
            "name": "BenchmarkEvict_200Files (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 8194,
            "unit": "B/op",
            "extra": "34266 times\n4 procs"
          },
          {
            "name": "BenchmarkEvict_200Files (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "34266 times\n4 procs"
          },
          {
            "name": "BenchmarkEvict_200Files (github.com/alex60217101990/terse/internal/cache)",
            "value": 36385,
            "unit": "ns/op\t    8194 B/op\t       1 allocs/op",
            "extra": "32760 times\n4 procs"
          },
          {
            "name": "BenchmarkEvict_200Files (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 36385,
            "unit": "ns/op",
            "extra": "32760 times\n4 procs"
          },
          {
            "name": "BenchmarkEvict_200Files (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 8194,
            "unit": "B/op",
            "extra": "32760 times\n4 procs"
          },
          {
            "name": "BenchmarkEvict_200Files (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "32760 times\n4 procs"
          },
          {
            "name": "BenchmarkEvict_200Files (github.com/alex60217101990/terse/internal/cache)",
            "value": 35574,
            "unit": "ns/op\t    8194 B/op\t       1 allocs/op",
            "extra": "33621 times\n4 procs"
          },
          {
            "name": "BenchmarkEvict_200Files (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 35574,
            "unit": "ns/op",
            "extra": "33621 times\n4 procs"
          },
          {
            "name": "BenchmarkEvict_200Files (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 8194,
            "unit": "B/op",
            "extra": "33621 times\n4 procs"
          },
          {
            "name": "BenchmarkEvict_200Files (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "33621 times\n4 procs"
          },
          {
            "name": "BenchmarkEvict_200Files (github.com/alex60217101990/terse/internal/cache)",
            "value": 35709,
            "unit": "ns/op\t    8194 B/op\t       1 allocs/op",
            "extra": "33835 times\n4 procs"
          },
          {
            "name": "BenchmarkEvict_200Files (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 35709,
            "unit": "ns/op",
            "extra": "33835 times\n4 procs"
          },
          {
            "name": "BenchmarkEvict_200Files (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 8194,
            "unit": "B/op",
            "extra": "33835 times\n4 procs"
          },
          {
            "name": "BenchmarkEvict_200Files (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "33835 times\n4 procs"
          },
          {
            "name": "BenchmarkRefPut (github.com/alex60217101990/terse/internal/cache)",
            "value": 35681,
            "unit": "ns/op\t    2520 B/op\t       9 allocs/op",
            "extra": "33921 times\n4 procs"
          },
          {
            "name": "BenchmarkRefPut (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 35681,
            "unit": "ns/op",
            "extra": "33921 times\n4 procs"
          },
          {
            "name": "BenchmarkRefPut (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 2520,
            "unit": "B/op",
            "extra": "33921 times\n4 procs"
          },
          {
            "name": "BenchmarkRefPut (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 9,
            "unit": "allocs/op",
            "extra": "33921 times\n4 procs"
          },
          {
            "name": "BenchmarkRefPut (github.com/alex60217101990/terse/internal/cache)",
            "value": 35401,
            "unit": "ns/op\t    2522 B/op\t       9 allocs/op",
            "extra": "33771 times\n4 procs"
          },
          {
            "name": "BenchmarkRefPut (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 35401,
            "unit": "ns/op",
            "extra": "33771 times\n4 procs"
          },
          {
            "name": "BenchmarkRefPut (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 2522,
            "unit": "B/op",
            "extra": "33771 times\n4 procs"
          },
          {
            "name": "BenchmarkRefPut (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 9,
            "unit": "allocs/op",
            "extra": "33771 times\n4 procs"
          },
          {
            "name": "BenchmarkRefPut (github.com/alex60217101990/terse/internal/cache)",
            "value": 35292,
            "unit": "ns/op\t    2524 B/op\t       9 allocs/op",
            "extra": "34118 times\n4 procs"
          },
          {
            "name": "BenchmarkRefPut (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 35292,
            "unit": "ns/op",
            "extra": "34118 times\n4 procs"
          },
          {
            "name": "BenchmarkRefPut (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 2524,
            "unit": "B/op",
            "extra": "34118 times\n4 procs"
          },
          {
            "name": "BenchmarkRefPut (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 9,
            "unit": "allocs/op",
            "extra": "34118 times\n4 procs"
          },
          {
            "name": "BenchmarkRefPut (github.com/alex60217101990/terse/internal/cache)",
            "value": 35219,
            "unit": "ns/op\t    2525 B/op\t       9 allocs/op",
            "extra": "33903 times\n4 procs"
          },
          {
            "name": "BenchmarkRefPut (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 35219,
            "unit": "ns/op",
            "extra": "33903 times\n4 procs"
          },
          {
            "name": "BenchmarkRefPut (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 2525,
            "unit": "B/op",
            "extra": "33903 times\n4 procs"
          },
          {
            "name": "BenchmarkRefPut (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 9,
            "unit": "allocs/op",
            "extra": "33903 times\n4 procs"
          },
          {
            "name": "BenchmarkRefPut (github.com/alex60217101990/terse/internal/cache)",
            "value": 35170,
            "unit": "ns/op\t    2507 B/op\t       9 allocs/op",
            "extra": "34256 times\n4 procs"
          },
          {
            "name": "BenchmarkRefPut (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 35170,
            "unit": "ns/op",
            "extra": "34256 times\n4 procs"
          },
          {
            "name": "BenchmarkRefPut (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 2507,
            "unit": "B/op",
            "extra": "34256 times\n4 procs"
          },
          {
            "name": "BenchmarkRefPut (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 9,
            "unit": "allocs/op",
            "extra": "34256 times\n4 procs"
          },
          {
            "name": "BenchmarkRefPut (github.com/alex60217101990/terse/internal/cache)",
            "value": 35109,
            "unit": "ns/op\t    2506 B/op\t       9 allocs/op",
            "extra": "34119 times\n4 procs"
          },
          {
            "name": "BenchmarkRefPut (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 35109,
            "unit": "ns/op",
            "extra": "34119 times\n4 procs"
          },
          {
            "name": "BenchmarkRefPut (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 2506,
            "unit": "B/op",
            "extra": "34119 times\n4 procs"
          },
          {
            "name": "BenchmarkRefPut (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 9,
            "unit": "allocs/op",
            "extra": "34119 times\n4 procs"
          },
          {
            "name": "BenchmarkRefGet (github.com/alex60217101990/terse/internal/cache)",
            "value": 12715,
            "unit": "ns/op\t    2133 B/op\t      11 allocs/op",
            "extra": "93460 times\n4 procs"
          },
          {
            "name": "BenchmarkRefGet (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 12715,
            "unit": "ns/op",
            "extra": "93460 times\n4 procs"
          },
          {
            "name": "BenchmarkRefGet (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 2133,
            "unit": "B/op",
            "extra": "93460 times\n4 procs"
          },
          {
            "name": "BenchmarkRefGet (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 11,
            "unit": "allocs/op",
            "extra": "93460 times\n4 procs"
          },
          {
            "name": "BenchmarkRefGet (github.com/alex60217101990/terse/internal/cache)",
            "value": 12597,
            "unit": "ns/op\t    2134 B/op\t      11 allocs/op",
            "extra": "94422 times\n4 procs"
          },
          {
            "name": "BenchmarkRefGet (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 12597,
            "unit": "ns/op",
            "extra": "94422 times\n4 procs"
          },
          {
            "name": "BenchmarkRefGet (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 2134,
            "unit": "B/op",
            "extra": "94422 times\n4 procs"
          },
          {
            "name": "BenchmarkRefGet (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 11,
            "unit": "allocs/op",
            "extra": "94422 times\n4 procs"
          },
          {
            "name": "BenchmarkRefGet (github.com/alex60217101990/terse/internal/cache)",
            "value": 12588,
            "unit": "ns/op\t    2133 B/op\t      11 allocs/op",
            "extra": "93735 times\n4 procs"
          },
          {
            "name": "BenchmarkRefGet (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 12588,
            "unit": "ns/op",
            "extra": "93735 times\n4 procs"
          },
          {
            "name": "BenchmarkRefGet (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 2133,
            "unit": "B/op",
            "extra": "93735 times\n4 procs"
          },
          {
            "name": "BenchmarkRefGet (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 11,
            "unit": "allocs/op",
            "extra": "93735 times\n4 procs"
          },
          {
            "name": "BenchmarkRefGet (github.com/alex60217101990/terse/internal/cache)",
            "value": 12591,
            "unit": "ns/op\t    2134 B/op\t      11 allocs/op",
            "extra": "95199 times\n4 procs"
          },
          {
            "name": "BenchmarkRefGet (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 12591,
            "unit": "ns/op",
            "extra": "95199 times\n4 procs"
          },
          {
            "name": "BenchmarkRefGet (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 2134,
            "unit": "B/op",
            "extra": "95199 times\n4 procs"
          },
          {
            "name": "BenchmarkRefGet (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 11,
            "unit": "allocs/op",
            "extra": "95199 times\n4 procs"
          },
          {
            "name": "BenchmarkRefGet (github.com/alex60217101990/terse/internal/cache)",
            "value": 12620,
            "unit": "ns/op\t    2118 B/op\t      11 allocs/op",
            "extra": "93919 times\n4 procs"
          },
          {
            "name": "BenchmarkRefGet (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 12620,
            "unit": "ns/op",
            "extra": "93919 times\n4 procs"
          },
          {
            "name": "BenchmarkRefGet (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 2118,
            "unit": "B/op",
            "extra": "93919 times\n4 procs"
          },
          {
            "name": "BenchmarkRefGet (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 11,
            "unit": "allocs/op",
            "extra": "93919 times\n4 procs"
          },
          {
            "name": "BenchmarkRefGet (github.com/alex60217101990/terse/internal/cache)",
            "value": 12535,
            "unit": "ns/op\t    2133 B/op\t      11 allocs/op",
            "extra": "95421 times\n4 procs"
          },
          {
            "name": "BenchmarkRefGet (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 12535,
            "unit": "ns/op",
            "extra": "95421 times\n4 procs"
          },
          {
            "name": "BenchmarkRefGet (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 2133,
            "unit": "B/op",
            "extra": "95421 times\n4 procs"
          },
          {
            "name": "BenchmarkRefGet (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 11,
            "unit": "allocs/op",
            "extra": "95421 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveLoad (github.com/alex60217101990/terse/internal/cache)",
            "value": 153693,
            "unit": "ns/op\t   15702 B/op\t      24 allocs/op",
            "extra": "8090 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveLoad (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 153693,
            "unit": "ns/op",
            "extra": "8090 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveLoad (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 15702,
            "unit": "B/op",
            "extra": "8090 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveLoad (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 24,
            "unit": "allocs/op",
            "extra": "8090 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveLoad (github.com/alex60217101990/terse/internal/cache)",
            "value": 150687,
            "unit": "ns/op\t   15702 B/op\t      24 allocs/op",
            "extra": "8344 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveLoad (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 150687,
            "unit": "ns/op",
            "extra": "8344 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveLoad (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 15702,
            "unit": "B/op",
            "extra": "8344 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveLoad (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 24,
            "unit": "allocs/op",
            "extra": "8344 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveLoad (github.com/alex60217101990/terse/internal/cache)",
            "value": 145583,
            "unit": "ns/op\t   15707 B/op\t      24 allocs/op",
            "extra": "7975 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveLoad (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 145583,
            "unit": "ns/op",
            "extra": "7975 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveLoad (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 15707,
            "unit": "B/op",
            "extra": "7975 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveLoad (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 24,
            "unit": "allocs/op",
            "extra": "7975 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveLoad (github.com/alex60217101990/terse/internal/cache)",
            "value": 147707,
            "unit": "ns/op\t   15708 B/op\t      24 allocs/op",
            "extra": "8671 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveLoad (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 147707,
            "unit": "ns/op",
            "extra": "8671 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveLoad (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 15708,
            "unit": "B/op",
            "extra": "8671 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveLoad (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 24,
            "unit": "allocs/op",
            "extra": "8671 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveLoad (github.com/alex60217101990/terse/internal/cache)",
            "value": 147797,
            "unit": "ns/op\t   15701 B/op\t      24 allocs/op",
            "extra": "8342 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveLoad (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 147797,
            "unit": "ns/op",
            "extra": "8342 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveLoad (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 15701,
            "unit": "B/op",
            "extra": "8342 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveLoad (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 24,
            "unit": "allocs/op",
            "extra": "8342 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveLoad (github.com/alex60217101990/terse/internal/cache)",
            "value": 145578,
            "unit": "ns/op\t   15688 B/op\t      24 allocs/op",
            "extra": "8136 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveLoad (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 145578,
            "unit": "ns/op",
            "extra": "8136 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveLoad (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 15688,
            "unit": "B/op",
            "extra": "8136 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveLoad (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 24,
            "unit": "allocs/op",
            "extra": "8136 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff_LargeFileSmallChange (github.com/alex60217101990/terse/internal/cache)",
            "value": 196302,
            "unit": "ns/op\t  299259 B/op\t      17 allocs/op",
            "extra": "6476 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff_LargeFileSmallChange (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 196302,
            "unit": "ns/op",
            "extra": "6476 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff_LargeFileSmallChange (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 299259,
            "unit": "B/op",
            "extra": "6476 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff_LargeFileSmallChange (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 17,
            "unit": "allocs/op",
            "extra": "6476 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff_LargeFileSmallChange (github.com/alex60217101990/terse/internal/cache)",
            "value": 187366,
            "unit": "ns/op\t  299171 B/op\t      17 allocs/op",
            "extra": "5756 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff_LargeFileSmallChange (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 187366,
            "unit": "ns/op",
            "extra": "5756 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff_LargeFileSmallChange (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 299171,
            "unit": "B/op",
            "extra": "5756 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff_LargeFileSmallChange (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 17,
            "unit": "allocs/op",
            "extra": "5756 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff_LargeFileSmallChange (github.com/alex60217101990/terse/internal/cache)",
            "value": 193715,
            "unit": "ns/op\t  299170 B/op\t      17 allocs/op",
            "extra": "6388 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff_LargeFileSmallChange (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 193715,
            "unit": "ns/op",
            "extra": "6388 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff_LargeFileSmallChange (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 299170,
            "unit": "B/op",
            "extra": "6388 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff_LargeFileSmallChange (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 17,
            "unit": "allocs/op",
            "extra": "6388 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff_LargeFileSmallChange (github.com/alex60217101990/terse/internal/cache)",
            "value": 200902,
            "unit": "ns/op\t  299171 B/op\t      17 allocs/op",
            "extra": "5970 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff_LargeFileSmallChange (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 200902,
            "unit": "ns/op",
            "extra": "5970 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff_LargeFileSmallChange (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 299171,
            "unit": "B/op",
            "extra": "5970 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff_LargeFileSmallChange (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 17,
            "unit": "allocs/op",
            "extra": "5970 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff_LargeFileSmallChange (github.com/alex60217101990/terse/internal/cache)",
            "value": 202685,
            "unit": "ns/op\t  299174 B/op\t      17 allocs/op",
            "extra": "6163 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff_LargeFileSmallChange (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 202685,
            "unit": "ns/op",
            "extra": "6163 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff_LargeFileSmallChange (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 299174,
            "unit": "B/op",
            "extra": "6163 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff_LargeFileSmallChange (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 17,
            "unit": "allocs/op",
            "extra": "6163 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff_LargeFileSmallChange (github.com/alex60217101990/terse/internal/cache)",
            "value": 203038,
            "unit": "ns/op\t  299174 B/op\t      17 allocs/op",
            "extra": "5418 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff_LargeFileSmallChange (github.com/alex60217101990/terse/internal/cache) - ns/op",
            "value": 203038,
            "unit": "ns/op",
            "extra": "5418 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff_LargeFileSmallChange (github.com/alex60217101990/terse/internal/cache) - B/op",
            "value": 299174,
            "unit": "B/op",
            "extra": "5418 times\n4 procs"
          },
          {
            "name": "BenchmarkUnifiedDiff_LargeFileSmallChange (github.com/alex60217101990/terse/internal/cache) - allocs/op",
            "value": 17,
            "unit": "allocs/op",
            "extra": "5418 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_ReadAll (github.com/alex60217101990/terse/internal/daemon)",
            "value": 4822,
            "unit": "ns/op\t   27824 B/op\t      12 allocs/op",
            "extra": "245670 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_ReadAll (github.com/alex60217101990/terse/internal/daemon) - ns/op",
            "value": 4822,
            "unit": "ns/op",
            "extra": "245670 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_ReadAll (github.com/alex60217101990/terse/internal/daemon) - B/op",
            "value": 27824,
            "unit": "B/op",
            "extra": "245670 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_ReadAll (github.com/alex60217101990/terse/internal/daemon) - allocs/op",
            "value": 12,
            "unit": "allocs/op",
            "extra": "245670 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_ReadAll (github.com/alex60217101990/terse/internal/daemon)",
            "value": 4749,
            "unit": "ns/op\t   27824 B/op\t      12 allocs/op",
            "extra": "251444 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_ReadAll (github.com/alex60217101990/terse/internal/daemon) - ns/op",
            "value": 4749,
            "unit": "ns/op",
            "extra": "251444 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_ReadAll (github.com/alex60217101990/terse/internal/daemon) - B/op",
            "value": 27824,
            "unit": "B/op",
            "extra": "251444 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_ReadAll (github.com/alex60217101990/terse/internal/daemon) - allocs/op",
            "value": 12,
            "unit": "allocs/op",
            "extra": "251444 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_ReadAll (github.com/alex60217101990/terse/internal/daemon)",
            "value": 4672,
            "unit": "ns/op\t   27824 B/op\t      12 allocs/op",
            "extra": "263137 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_ReadAll (github.com/alex60217101990/terse/internal/daemon) - ns/op",
            "value": 4672,
            "unit": "ns/op",
            "extra": "263137 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_ReadAll (github.com/alex60217101990/terse/internal/daemon) - B/op",
            "value": 27824,
            "unit": "B/op",
            "extra": "263137 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_ReadAll (github.com/alex60217101990/terse/internal/daemon) - allocs/op",
            "value": 12,
            "unit": "allocs/op",
            "extra": "263137 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_ReadAll (github.com/alex60217101990/terse/internal/daemon)",
            "value": 4999,
            "unit": "ns/op\t   27824 B/op\t      12 allocs/op",
            "extra": "264127 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_ReadAll (github.com/alex60217101990/terse/internal/daemon) - ns/op",
            "value": 4999,
            "unit": "ns/op",
            "extra": "264127 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_ReadAll (github.com/alex60217101990/terse/internal/daemon) - B/op",
            "value": 27824,
            "unit": "B/op",
            "extra": "264127 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_ReadAll (github.com/alex60217101990/terse/internal/daemon) - allocs/op",
            "value": 12,
            "unit": "allocs/op",
            "extra": "264127 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_ReadAll (github.com/alex60217101990/terse/internal/daemon)",
            "value": 4746,
            "unit": "ns/op\t   27824 B/op\t      12 allocs/op",
            "extra": "255166 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_ReadAll (github.com/alex60217101990/terse/internal/daemon) - ns/op",
            "value": 4746,
            "unit": "ns/op",
            "extra": "255166 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_ReadAll (github.com/alex60217101990/terse/internal/daemon) - B/op",
            "value": 27824,
            "unit": "B/op",
            "extra": "255166 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_ReadAll (github.com/alex60217101990/terse/internal/daemon) - allocs/op",
            "value": 12,
            "unit": "allocs/op",
            "extra": "255166 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_ReadAll (github.com/alex60217101990/terse/internal/daemon)",
            "value": 4732,
            "unit": "ns/op\t   27824 B/op\t      12 allocs/op",
            "extra": "222692 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_ReadAll (github.com/alex60217101990/terse/internal/daemon) - ns/op",
            "value": 4732,
            "unit": "ns/op",
            "extra": "222692 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_ReadAll (github.com/alex60217101990/terse/internal/daemon) - B/op",
            "value": 27824,
            "unit": "B/op",
            "extra": "222692 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_ReadAll (github.com/alex60217101990/terse/internal/daemon) - allocs/op",
            "value": 12,
            "unit": "allocs/op",
            "extra": "222692 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_Pooled (github.com/alex60217101990/terse/internal/daemon)",
            "value": 201.7,
            "unit": "ns/op\t      48 B/op\t       1 allocs/op",
            "extra": "5997872 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_Pooled (github.com/alex60217101990/terse/internal/daemon) - ns/op",
            "value": 201.7,
            "unit": "ns/op",
            "extra": "5997872 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_Pooled (github.com/alex60217101990/terse/internal/daemon) - B/op",
            "value": 48,
            "unit": "B/op",
            "extra": "5997872 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_Pooled (github.com/alex60217101990/terse/internal/daemon) - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "5997872 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_Pooled (github.com/alex60217101990/terse/internal/daemon)",
            "value": 202.7,
            "unit": "ns/op\t      48 B/op\t       1 allocs/op",
            "extra": "5917051 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_Pooled (github.com/alex60217101990/terse/internal/daemon) - ns/op",
            "value": 202.7,
            "unit": "ns/op",
            "extra": "5917051 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_Pooled (github.com/alex60217101990/terse/internal/daemon) - B/op",
            "value": 48,
            "unit": "B/op",
            "extra": "5917051 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_Pooled (github.com/alex60217101990/terse/internal/daemon) - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "5917051 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_Pooled (github.com/alex60217101990/terse/internal/daemon)",
            "value": 202.3,
            "unit": "ns/op\t      48 B/op\t       1 allocs/op",
            "extra": "5939209 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_Pooled (github.com/alex60217101990/terse/internal/daemon) - ns/op",
            "value": 202.3,
            "unit": "ns/op",
            "extra": "5939209 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_Pooled (github.com/alex60217101990/terse/internal/daemon) - B/op",
            "value": 48,
            "unit": "B/op",
            "extra": "5939209 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_Pooled (github.com/alex60217101990/terse/internal/daemon) - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "5939209 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_Pooled (github.com/alex60217101990/terse/internal/daemon)",
            "value": 202,
            "unit": "ns/op\t      48 B/op\t       1 allocs/op",
            "extra": "5964672 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_Pooled (github.com/alex60217101990/terse/internal/daemon) - ns/op",
            "value": 202,
            "unit": "ns/op",
            "extra": "5964672 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_Pooled (github.com/alex60217101990/terse/internal/daemon) - B/op",
            "value": 48,
            "unit": "B/op",
            "extra": "5964672 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_Pooled (github.com/alex60217101990/terse/internal/daemon) - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "5964672 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_Pooled (github.com/alex60217101990/terse/internal/daemon)",
            "value": 203.3,
            "unit": "ns/op\t      48 B/op\t       1 allocs/op",
            "extra": "5896038 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_Pooled (github.com/alex60217101990/terse/internal/daemon) - ns/op",
            "value": 203.3,
            "unit": "ns/op",
            "extra": "5896038 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_Pooled (github.com/alex60217101990/terse/internal/daemon) - B/op",
            "value": 48,
            "unit": "B/op",
            "extra": "5896038 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_Pooled (github.com/alex60217101990/terse/internal/daemon) - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "5896038 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_Pooled (github.com/alex60217101990/terse/internal/daemon)",
            "value": 201.6,
            "unit": "ns/op\t      48 B/op\t       1 allocs/op",
            "extra": "5984366 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_Pooled (github.com/alex60217101990/terse/internal/daemon) - ns/op",
            "value": 201.6,
            "unit": "ns/op",
            "extra": "5984366 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_Pooled (github.com/alex60217101990/terse/internal/daemon) - B/op",
            "value": 48,
            "unit": "B/op",
            "extra": "5984366 times\n4 procs"
          },
          {
            "name": "BenchmarkReadPath_Pooled (github.com/alex60217101990/terse/internal/daemon) - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "5984366 times\n4 procs"
          },
          {
            "name": "BenchmarkDaemonRoundtrip (github.com/alex60217101990/terse/internal/daemon)",
            "value": 128803,
            "unit": "ns/op\t    6247 B/op\t      55 allocs/op",
            "extra": "8796 times\n4 procs"
          },
          {
            "name": "BenchmarkDaemonRoundtrip (github.com/alex60217101990/terse/internal/daemon) - ns/op",
            "value": 128803,
            "unit": "ns/op",
            "extra": "8796 times\n4 procs"
          },
          {
            "name": "BenchmarkDaemonRoundtrip (github.com/alex60217101990/terse/internal/daemon) - B/op",
            "value": 6247,
            "unit": "B/op",
            "extra": "8796 times\n4 procs"
          },
          {
            "name": "BenchmarkDaemonRoundtrip (github.com/alex60217101990/terse/internal/daemon) - allocs/op",
            "value": 55,
            "unit": "allocs/op",
            "extra": "8796 times\n4 procs"
          },
          {
            "name": "BenchmarkDaemonRoundtrip (github.com/alex60217101990/terse/internal/daemon)",
            "value": 128943,
            "unit": "ns/op\t    6231 B/op\t      55 allocs/op",
            "extra": "8950 times\n4 procs"
          },
          {
            "name": "BenchmarkDaemonRoundtrip (github.com/alex60217101990/terse/internal/daemon) - ns/op",
            "value": 128943,
            "unit": "ns/op",
            "extra": "8950 times\n4 procs"
          },
          {
            "name": "BenchmarkDaemonRoundtrip (github.com/alex60217101990/terse/internal/daemon) - B/op",
            "value": 6231,
            "unit": "B/op",
            "extra": "8950 times\n4 procs"
          },
          {
            "name": "BenchmarkDaemonRoundtrip (github.com/alex60217101990/terse/internal/daemon) - allocs/op",
            "value": 55,
            "unit": "allocs/op",
            "extra": "8950 times\n4 procs"
          },
          {
            "name": "BenchmarkDaemonRoundtrip (github.com/alex60217101990/terse/internal/daemon)",
            "value": 129208,
            "unit": "ns/op\t    6248 B/op\t      55 allocs/op",
            "extra": "9130 times\n4 procs"
          },
          {
            "name": "BenchmarkDaemonRoundtrip (github.com/alex60217101990/terse/internal/daemon) - ns/op",
            "value": 129208,
            "unit": "ns/op",
            "extra": "9130 times\n4 procs"
          },
          {
            "name": "BenchmarkDaemonRoundtrip (github.com/alex60217101990/terse/internal/daemon) - B/op",
            "value": 6248,
            "unit": "B/op",
            "extra": "9130 times\n4 procs"
          },
          {
            "name": "BenchmarkDaemonRoundtrip (github.com/alex60217101990/terse/internal/daemon) - allocs/op",
            "value": 55,
            "unit": "allocs/op",
            "extra": "9130 times\n4 procs"
          },
          {
            "name": "BenchmarkDaemonRoundtrip (github.com/alex60217101990/terse/internal/daemon)",
            "value": 128931,
            "unit": "ns/op\t    6251 B/op\t      55 allocs/op",
            "extra": "8728 times\n4 procs"
          },
          {
            "name": "BenchmarkDaemonRoundtrip (github.com/alex60217101990/terse/internal/daemon) - ns/op",
            "value": 128931,
            "unit": "ns/op",
            "extra": "8728 times\n4 procs"
          },
          {
            "name": "BenchmarkDaemonRoundtrip (github.com/alex60217101990/terse/internal/daemon) - B/op",
            "value": 6251,
            "unit": "B/op",
            "extra": "8728 times\n4 procs"
          },
          {
            "name": "BenchmarkDaemonRoundtrip (github.com/alex60217101990/terse/internal/daemon) - allocs/op",
            "value": 55,
            "unit": "allocs/op",
            "extra": "8728 times\n4 procs"
          },
          {
            "name": "BenchmarkDaemonRoundtrip (github.com/alex60217101990/terse/internal/daemon)",
            "value": 128368,
            "unit": "ns/op\t    6254 B/op\t      55 allocs/op",
            "extra": "8995 times\n4 procs"
          },
          {
            "name": "BenchmarkDaemonRoundtrip (github.com/alex60217101990/terse/internal/daemon) - ns/op",
            "value": 128368,
            "unit": "ns/op",
            "extra": "8995 times\n4 procs"
          },
          {
            "name": "BenchmarkDaemonRoundtrip (github.com/alex60217101990/terse/internal/daemon) - B/op",
            "value": 6254,
            "unit": "B/op",
            "extra": "8995 times\n4 procs"
          },
          {
            "name": "BenchmarkDaemonRoundtrip (github.com/alex60217101990/terse/internal/daemon) - allocs/op",
            "value": 55,
            "unit": "allocs/op",
            "extra": "8995 times\n4 procs"
          },
          {
            "name": "BenchmarkDaemonRoundtrip (github.com/alex60217101990/terse/internal/daemon)",
            "value": 127659,
            "unit": "ns/op\t    6222 B/op\t      55 allocs/op",
            "extra": "9423 times\n4 procs"
          },
          {
            "name": "BenchmarkDaemonRoundtrip (github.com/alex60217101990/terse/internal/daemon) - ns/op",
            "value": 127659,
            "unit": "ns/op",
            "extra": "9423 times\n4 procs"
          },
          {
            "name": "BenchmarkDaemonRoundtrip (github.com/alex60217101990/terse/internal/daemon) - B/op",
            "value": 6222,
            "unit": "B/op",
            "extra": "9423 times\n4 procs"
          },
          {
            "name": "BenchmarkDaemonRoundtrip (github.com/alex60217101990/terse/internal/daemon) - allocs/op",
            "value": 55,
            "unit": "allocs/op",
            "extra": "9423 times\n4 procs"
          },
          {
            "name": "BenchmarkCLIRoundtrip (github.com/alex60217101990/terse/internal/daemon)",
            "value": 3150652,
            "unit": "ns/op\t    4198 B/op\t      55 allocs/op",
            "extra": "387 times\n4 procs"
          },
          {
            "name": "BenchmarkCLIRoundtrip (github.com/alex60217101990/terse/internal/daemon) - ns/op",
            "value": 3150652,
            "unit": "ns/op",
            "extra": "387 times\n4 procs"
          },
          {
            "name": "BenchmarkCLIRoundtrip (github.com/alex60217101990/terse/internal/daemon) - B/op",
            "value": 4198,
            "unit": "B/op",
            "extra": "387 times\n4 procs"
          },
          {
            "name": "BenchmarkCLIRoundtrip (github.com/alex60217101990/terse/internal/daemon) - allocs/op",
            "value": 55,
            "unit": "allocs/op",
            "extra": "387 times\n4 procs"
          },
          {
            "name": "BenchmarkCLIRoundtrip (github.com/alex60217101990/terse/internal/daemon)",
            "value": 3047182,
            "unit": "ns/op\t    4095 B/op\t      55 allocs/op",
            "extra": "393 times\n4 procs"
          },
          {
            "name": "BenchmarkCLIRoundtrip (github.com/alex60217101990/terse/internal/daemon) - ns/op",
            "value": 3047182,
            "unit": "ns/op",
            "extra": "393 times\n4 procs"
          },
          {
            "name": "BenchmarkCLIRoundtrip (github.com/alex60217101990/terse/internal/daemon) - B/op",
            "value": 4095,
            "unit": "B/op",
            "extra": "393 times\n4 procs"
          },
          {
            "name": "BenchmarkCLIRoundtrip (github.com/alex60217101990/terse/internal/daemon) - allocs/op",
            "value": 55,
            "unit": "allocs/op",
            "extra": "393 times\n4 procs"
          },
          {
            "name": "BenchmarkCLIRoundtrip (github.com/alex60217101990/terse/internal/daemon)",
            "value": 3085049,
            "unit": "ns/op\t    4176 B/op\t      55 allocs/op",
            "extra": "388 times\n4 procs"
          },
          {
            "name": "BenchmarkCLIRoundtrip (github.com/alex60217101990/terse/internal/daemon) - ns/op",
            "value": 3085049,
            "unit": "ns/op",
            "extra": "388 times\n4 procs"
          },
          {
            "name": "BenchmarkCLIRoundtrip (github.com/alex60217101990/terse/internal/daemon) - B/op",
            "value": 4176,
            "unit": "B/op",
            "extra": "388 times\n4 procs"
          },
          {
            "name": "BenchmarkCLIRoundtrip (github.com/alex60217101990/terse/internal/daemon) - allocs/op",
            "value": 55,
            "unit": "allocs/op",
            "extra": "388 times\n4 procs"
          },
          {
            "name": "BenchmarkCLIRoundtrip (github.com/alex60217101990/terse/internal/daemon)",
            "value": 3093162,
            "unit": "ns/op\t    4132 B/op\t      55 allocs/op",
            "extra": "390 times\n4 procs"
          },
          {
            "name": "BenchmarkCLIRoundtrip (github.com/alex60217101990/terse/internal/daemon) - ns/op",
            "value": 3093162,
            "unit": "ns/op",
            "extra": "390 times\n4 procs"
          },
          {
            "name": "BenchmarkCLIRoundtrip (github.com/alex60217101990/terse/internal/daemon) - B/op",
            "value": 4132,
            "unit": "B/op",
            "extra": "390 times\n4 procs"
          },
          {
            "name": "BenchmarkCLIRoundtrip (github.com/alex60217101990/terse/internal/daemon) - allocs/op",
            "value": 55,
            "unit": "allocs/op",
            "extra": "390 times\n4 procs"
          },
          {
            "name": "BenchmarkCLIRoundtrip (github.com/alex60217101990/terse/internal/daemon)",
            "value": 3074548,
            "unit": "ns/op\t    4085 B/op\t      55 allocs/op",
            "extra": "397 times\n4 procs"
          },
          {
            "name": "BenchmarkCLIRoundtrip (github.com/alex60217101990/terse/internal/daemon) - ns/op",
            "value": 3074548,
            "unit": "ns/op",
            "extra": "397 times\n4 procs"
          },
          {
            "name": "BenchmarkCLIRoundtrip (github.com/alex60217101990/terse/internal/daemon) - B/op",
            "value": 4085,
            "unit": "B/op",
            "extra": "397 times\n4 procs"
          },
          {
            "name": "BenchmarkCLIRoundtrip (github.com/alex60217101990/terse/internal/daemon) - allocs/op",
            "value": 55,
            "unit": "allocs/op",
            "extra": "397 times\n4 procs"
          },
          {
            "name": "BenchmarkCLIRoundtrip (github.com/alex60217101990/terse/internal/daemon)",
            "value": 3095480,
            "unit": "ns/op\t    4086 B/op\t      55 allocs/op",
            "extra": "386 times\n4 procs"
          },
          {
            "name": "BenchmarkCLIRoundtrip (github.com/alex60217101990/terse/internal/daemon) - ns/op",
            "value": 3095480,
            "unit": "ns/op",
            "extra": "386 times\n4 procs"
          },
          {
            "name": "BenchmarkCLIRoundtrip (github.com/alex60217101990/terse/internal/daemon) - B/op",
            "value": 4086,
            "unit": "B/op",
            "extra": "386 times\n4 procs"
          },
          {
            "name": "BenchmarkCLIRoundtrip (github.com/alex60217101990/terse/internal/daemon) - allocs/op",
            "value": 55,
            "unit": "allocs/op",
            "extra": "386 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_NoDup (github.com/alex60217101990/terse/internal/detect)",
            "value": 1047,
            "unit": "ns/op\t 779.29 MB/s\t       0 B/op\t       0 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_NoDup (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 1047,
            "unit": "ns/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_NoDup (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 779.29,
            "unit": "MB/s",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_NoDup (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_NoDup (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_NoDup (github.com/alex60217101990/terse/internal/detect)",
            "value": 1053,
            "unit": "ns/op\t 774.75 MB/s\t       0 B/op\t       0 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_NoDup (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 1053,
            "unit": "ns/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_NoDup (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 774.75,
            "unit": "MB/s",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_NoDup (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_NoDup (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_NoDup (github.com/alex60217101990/terse/internal/detect)",
            "value": 1046,
            "unit": "ns/op\t 779.84 MB/s\t       0 B/op\t       0 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_NoDup (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 1046,
            "unit": "ns/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_NoDup (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 779.84,
            "unit": "MB/s",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_NoDup (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_NoDup (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_NoDup (github.com/alex60217101990/terse/internal/detect)",
            "value": 1051,
            "unit": "ns/op\t 776.29 MB/s\t       0 B/op\t       0 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_NoDup (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 1051,
            "unit": "ns/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_NoDup (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 776.29,
            "unit": "MB/s",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_NoDup (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_NoDup (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_NoDup (github.com/alex60217101990/terse/internal/detect)",
            "value": 1053,
            "unit": "ns/op\t 774.58 MB/s\t       0 B/op\t       0 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_NoDup (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 1053,
            "unit": "ns/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_NoDup (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 774.58,
            "unit": "MB/s",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_NoDup (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_NoDup (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_NoDup (github.com/alex60217101990/terse/internal/detect)",
            "value": 1070,
            "unit": "ns/op\t 762.39 MB/s\t       0 B/op\t       0 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_NoDup (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 1070,
            "unit": "ns/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_NoDup (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 762.39,
            "unit": "MB/s",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_NoDup (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_NoDup (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks (github.com/alex60217101990/terse/internal/detect)",
            "value": 1425,
            "unit": "ns/op\t 805.36 MB/s\t    1152 B/op\t       1 allocs/op",
            "extra": "837613 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 1425,
            "unit": "ns/op",
            "extra": "837613 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 805.36,
            "unit": "MB/s",
            "extra": "837613 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 1152,
            "unit": "B/op",
            "extra": "837613 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "837613 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks (github.com/alex60217101990/terse/internal/detect)",
            "value": 1445,
            "unit": "ns/op\t 794.51 MB/s\t    1152 B/op\t       1 allocs/op",
            "extra": "848900 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 1445,
            "unit": "ns/op",
            "extra": "848900 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 794.51,
            "unit": "MB/s",
            "extra": "848900 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 1152,
            "unit": "B/op",
            "extra": "848900 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "848900 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks (github.com/alex60217101990/terse/internal/detect)",
            "value": 1447,
            "unit": "ns/op\t 793.55 MB/s\t    1152 B/op\t       1 allocs/op",
            "extra": "814905 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 1447,
            "unit": "ns/op",
            "extra": "814905 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 793.55,
            "unit": "MB/s",
            "extra": "814905 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 1152,
            "unit": "B/op",
            "extra": "814905 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "814905 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks (github.com/alex60217101990/terse/internal/detect)",
            "value": 1450,
            "unit": "ns/op\t 791.98 MB/s\t    1152 B/op\t       1 allocs/op",
            "extra": "842593 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 1450,
            "unit": "ns/op",
            "extra": "842593 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 791.98,
            "unit": "MB/s",
            "extra": "842593 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 1152,
            "unit": "B/op",
            "extra": "842593 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "842593 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks (github.com/alex60217101990/terse/internal/detect)",
            "value": 1445,
            "unit": "ns/op\t 794.72 MB/s\t    1152 B/op\t       1 allocs/op",
            "extra": "789196 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 1445,
            "unit": "ns/op",
            "extra": "789196 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 794.72,
            "unit": "MB/s",
            "extra": "789196 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 1152,
            "unit": "B/op",
            "extra": "789196 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "789196 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks (github.com/alex60217101990/terse/internal/detect)",
            "value": 1436,
            "unit": "ns/op\t 799.63 MB/s\t    1152 B/op\t       1 allocs/op",
            "extra": "846766 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 1436,
            "unit": "ns/op",
            "extra": "846766 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 799.63,
            "unit": "MB/s",
            "extra": "846766 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 1152,
            "unit": "B/op",
            "extra": "846766 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "846766 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_Fuzzy (github.com/alex60217101990/terse/internal/detect)",
            "value": 3705,
            "unit": "ns/op\t  94.74 MB/s\t    1008 B/op\t       6 allocs/op",
            "extra": "311143 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_Fuzzy (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 3705,
            "unit": "ns/op",
            "extra": "311143 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_Fuzzy (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 94.74,
            "unit": "MB/s",
            "extra": "311143 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_Fuzzy (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 1008,
            "unit": "B/op",
            "extra": "311143 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_Fuzzy (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 6,
            "unit": "allocs/op",
            "extra": "311143 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_Fuzzy (github.com/alex60217101990/terse/internal/detect)",
            "value": 3698,
            "unit": "ns/op\t  94.92 MB/s\t    1008 B/op\t       6 allocs/op",
            "extra": "319744 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_Fuzzy (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 3698,
            "unit": "ns/op",
            "extra": "319744 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_Fuzzy (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 94.92,
            "unit": "MB/s",
            "extra": "319744 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_Fuzzy (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 1008,
            "unit": "B/op",
            "extra": "319744 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_Fuzzy (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 6,
            "unit": "allocs/op",
            "extra": "319744 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_Fuzzy (github.com/alex60217101990/terse/internal/detect)",
            "value": 3706,
            "unit": "ns/op\t  94.70 MB/s\t    1008 B/op\t       6 allocs/op",
            "extra": "315783 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_Fuzzy (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 3706,
            "unit": "ns/op",
            "extra": "315783 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_Fuzzy (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 94.7,
            "unit": "MB/s",
            "extra": "315783 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_Fuzzy (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 1008,
            "unit": "B/op",
            "extra": "315783 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_Fuzzy (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 6,
            "unit": "allocs/op",
            "extra": "315783 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_Fuzzy (github.com/alex60217101990/terse/internal/detect)",
            "value": 3700,
            "unit": "ns/op\t  94.86 MB/s\t    1008 B/op\t       6 allocs/op",
            "extra": "323457 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_Fuzzy (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 3700,
            "unit": "ns/op",
            "extra": "323457 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_Fuzzy (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 94.86,
            "unit": "MB/s",
            "extra": "323457 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_Fuzzy (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 1008,
            "unit": "B/op",
            "extra": "323457 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_Fuzzy (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 6,
            "unit": "allocs/op",
            "extra": "323457 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_Fuzzy (github.com/alex60217101990/terse/internal/detect)",
            "value": 3712,
            "unit": "ns/op\t  94.57 MB/s\t    1008 B/op\t       6 allocs/op",
            "extra": "323757 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_Fuzzy (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 3712,
            "unit": "ns/op",
            "extra": "323757 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_Fuzzy (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 94.57,
            "unit": "MB/s",
            "extra": "323757 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_Fuzzy (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 1008,
            "unit": "B/op",
            "extra": "323757 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_Fuzzy (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 6,
            "unit": "allocs/op",
            "extra": "323757 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_Fuzzy (github.com/alex60217101990/terse/internal/detect)",
            "value": 3764,
            "unit": "ns/op\t  93.24 MB/s\t    1008 B/op\t       6 allocs/op",
            "extra": "324709 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_Fuzzy (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 3764,
            "unit": "ns/op",
            "extra": "324709 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_Fuzzy (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 93.24,
            "unit": "MB/s",
            "extra": "324709 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_Fuzzy (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 1008,
            "unit": "B/op",
            "extra": "324709 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldRepeatedBlocks_Fuzzy (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 6,
            "unit": "allocs/op",
            "extra": "324709 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff (github.com/alex60217101990/terse/internal/detect)",
            "value": 808.2,
            "unit": "ns/op\t2120.77 MB/s\t    1792 B/op\t       1 allocs/op",
            "extra": "1484966 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 808.2,
            "unit": "ns/op",
            "extra": "1484966 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 2120.77,
            "unit": "MB/s",
            "extra": "1484966 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 1792,
            "unit": "B/op",
            "extra": "1484966 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "1484966 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff (github.com/alex60217101990/terse/internal/detect)",
            "value": 838.4,
            "unit": "ns/op\t2044.44 MB/s\t    1792 B/op\t       1 allocs/op",
            "extra": "1458818 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 838.4,
            "unit": "ns/op",
            "extra": "1458818 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 2044.44,
            "unit": "MB/s",
            "extra": "1458818 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 1792,
            "unit": "B/op",
            "extra": "1458818 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "1458818 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff (github.com/alex60217101990/terse/internal/detect)",
            "value": 808.3,
            "unit": "ns/op\t2120.46 MB/s\t    1792 B/op\t       1 allocs/op",
            "extra": "1482474 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 808.3,
            "unit": "ns/op",
            "extra": "1482474 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 2120.46,
            "unit": "MB/s",
            "extra": "1482474 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 1792,
            "unit": "B/op",
            "extra": "1482474 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "1482474 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff (github.com/alex60217101990/terse/internal/detect)",
            "value": 803.7,
            "unit": "ns/op\t2132.66 MB/s\t    1792 B/op\t       1 allocs/op",
            "extra": "1497459 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 803.7,
            "unit": "ns/op",
            "extra": "1497459 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 2132.66,
            "unit": "MB/s",
            "extra": "1497459 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 1792,
            "unit": "B/op",
            "extra": "1497459 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "1497459 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff (github.com/alex60217101990/terse/internal/detect)",
            "value": 804.8,
            "unit": "ns/op\t2129.60 MB/s\t    1792 B/op\t       1 allocs/op",
            "extra": "1490308 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 804.8,
            "unit": "ns/op",
            "extra": "1490308 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 2129.6,
            "unit": "MB/s",
            "extra": "1490308 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 1792,
            "unit": "B/op",
            "extra": "1490308 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "1490308 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff (github.com/alex60217101990/terse/internal/detect)",
            "value": 805.7,
            "unit": "ns/op\t2127.40 MB/s\t    1792 B/op\t       1 allocs/op",
            "extra": "1499397 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 805.7,
            "unit": "ns/op",
            "extra": "1499397 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 2127.4,
            "unit": "MB/s",
            "extra": "1499397 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 1792,
            "unit": "B/op",
            "extra": "1499397 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "1499397 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff_NonMatch (github.com/alex60217101990/terse/internal/detect)",
            "value": 313.7,
            "unit": "ns/op\t9562.43 MB/s\t       0 B/op\t       0 allocs/op",
            "extra": "3836974 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff_NonMatch (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 313.7,
            "unit": "ns/op",
            "extra": "3836974 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff_NonMatch (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 9562.43,
            "unit": "MB/s",
            "extra": "3836974 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff_NonMatch (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "3836974 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff_NonMatch (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "3836974 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff_NonMatch (github.com/alex60217101990/terse/internal/detect)",
            "value": 312.7,
            "unit": "ns/op\t9593.43 MB/s\t       0 B/op\t       0 allocs/op",
            "extra": "3826222 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff_NonMatch (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 312.7,
            "unit": "ns/op",
            "extra": "3826222 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff_NonMatch (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 9593.43,
            "unit": "MB/s",
            "extra": "3826222 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff_NonMatch (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "3826222 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff_NonMatch (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "3826222 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff_NonMatch (github.com/alex60217101990/terse/internal/detect)",
            "value": 312.6,
            "unit": "ns/op\t9595.44 MB/s\t       0 B/op\t       0 allocs/op",
            "extra": "3836679 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff_NonMatch (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 312.6,
            "unit": "ns/op",
            "extra": "3836679 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff_NonMatch (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 9595.44,
            "unit": "MB/s",
            "extra": "3836679 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff_NonMatch (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "3836679 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff_NonMatch (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "3836679 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff_NonMatch (github.com/alex60217101990/terse/internal/detect)",
            "value": 312.6,
            "unit": "ns/op\t9598.46 MB/s\t       0 B/op\t       0 allocs/op",
            "extra": "3839803 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff_NonMatch (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 312.6,
            "unit": "ns/op",
            "extra": "3839803 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff_NonMatch (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 9598.46,
            "unit": "MB/s",
            "extra": "3839803 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff_NonMatch (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "3839803 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff_NonMatch (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "3839803 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff_NonMatch (github.com/alex60217101990/terse/internal/detect)",
            "value": 312.8,
            "unit": "ns/op\t9592.19 MB/s\t       0 B/op\t       0 allocs/op",
            "extra": "3834262 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff_NonMatch (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 312.8,
            "unit": "ns/op",
            "extra": "3834262 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff_NonMatch (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 9592.19,
            "unit": "MB/s",
            "extra": "3834262 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff_NonMatch (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "3834262 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff_NonMatch (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "3834262 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff_NonMatch (github.com/alex60217101990/terse/internal/detect)",
            "value": 312.9,
            "unit": "ns/op\t9588.04 MB/s\t       0 B/op\t       0 allocs/op",
            "extra": "3822552 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff_NonMatch (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 312.9,
            "unit": "ns/op",
            "extra": "3822552 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff_NonMatch (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 9588.04,
            "unit": "MB/s",
            "extra": "3822552 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff_NonMatch (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "3822552 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeGitDiff_NonMatch (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "3822552 times\n4 procs"
          },
          {
            "name": "BenchmarkAnalyzeJSONArray (github.com/alex60217101990/terse/internal/detect)",
            "value": 529332,
            "unit": "ns/op\t   18768 B/op\t      44 allocs/op",
            "extra": "2215 times\n4 procs"
          },
          {
            "name": "BenchmarkAnalyzeJSONArray (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 529332,
            "unit": "ns/op",
            "extra": "2215 times\n4 procs"
          },
          {
            "name": "BenchmarkAnalyzeJSONArray (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 18768,
            "unit": "B/op",
            "extra": "2215 times\n4 procs"
          },
          {
            "name": "BenchmarkAnalyzeJSONArray (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 44,
            "unit": "allocs/op",
            "extra": "2215 times\n4 procs"
          },
          {
            "name": "BenchmarkAnalyzeJSONArray (github.com/alex60217101990/terse/internal/detect)",
            "value": 527236,
            "unit": "ns/op\t   18768 B/op\t      44 allocs/op",
            "extra": "2278 times\n4 procs"
          },
          {
            "name": "BenchmarkAnalyzeJSONArray (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 527236,
            "unit": "ns/op",
            "extra": "2278 times\n4 procs"
          },
          {
            "name": "BenchmarkAnalyzeJSONArray (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 18768,
            "unit": "B/op",
            "extra": "2278 times\n4 procs"
          },
          {
            "name": "BenchmarkAnalyzeJSONArray (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 44,
            "unit": "allocs/op",
            "extra": "2278 times\n4 procs"
          },
          {
            "name": "BenchmarkAnalyzeJSONArray (github.com/alex60217101990/terse/internal/detect)",
            "value": 528463,
            "unit": "ns/op\t   18768 B/op\t      44 allocs/op",
            "extra": "2281 times\n4 procs"
          },
          {
            "name": "BenchmarkAnalyzeJSONArray (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 528463,
            "unit": "ns/op",
            "extra": "2281 times\n4 procs"
          },
          {
            "name": "BenchmarkAnalyzeJSONArray (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 18768,
            "unit": "B/op",
            "extra": "2281 times\n4 procs"
          },
          {
            "name": "BenchmarkAnalyzeJSONArray (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 44,
            "unit": "allocs/op",
            "extra": "2281 times\n4 procs"
          },
          {
            "name": "BenchmarkAnalyzeJSONArray (github.com/alex60217101990/terse/internal/detect)",
            "value": 528818,
            "unit": "ns/op\t   18768 B/op\t      44 allocs/op",
            "extra": "2271 times\n4 procs"
          },
          {
            "name": "BenchmarkAnalyzeJSONArray (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 528818,
            "unit": "ns/op",
            "extra": "2271 times\n4 procs"
          },
          {
            "name": "BenchmarkAnalyzeJSONArray (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 18768,
            "unit": "B/op",
            "extra": "2271 times\n4 procs"
          },
          {
            "name": "BenchmarkAnalyzeJSONArray (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 44,
            "unit": "allocs/op",
            "extra": "2271 times\n4 procs"
          },
          {
            "name": "BenchmarkAnalyzeJSONArray (github.com/alex60217101990/terse/internal/detect)",
            "value": 527290,
            "unit": "ns/op\t   18768 B/op\t      44 allocs/op",
            "extra": "2270 times\n4 procs"
          },
          {
            "name": "BenchmarkAnalyzeJSONArray (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 527290,
            "unit": "ns/op",
            "extra": "2270 times\n4 procs"
          },
          {
            "name": "BenchmarkAnalyzeJSONArray (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 18768,
            "unit": "B/op",
            "extra": "2270 times\n4 procs"
          },
          {
            "name": "BenchmarkAnalyzeJSONArray (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 44,
            "unit": "allocs/op",
            "extra": "2270 times\n4 procs"
          },
          {
            "name": "BenchmarkAnalyzeJSONArray (github.com/alex60217101990/terse/internal/detect)",
            "value": 527197,
            "unit": "ns/op\t   18768 B/op\t      44 allocs/op",
            "extra": "2268 times\n4 procs"
          },
          {
            "name": "BenchmarkAnalyzeJSONArray (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 527197,
            "unit": "ns/op",
            "extra": "2268 times\n4 procs"
          },
          {
            "name": "BenchmarkAnalyzeJSONArray (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 18768,
            "unit": "B/op",
            "extra": "2268 times\n4 procs"
          },
          {
            "name": "BenchmarkAnalyzeJSONArray (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 44,
            "unit": "allocs/op",
            "extra": "2268 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_Match (github.com/alex60217101990/terse/internal/detect)",
            "value": 37452,
            "unit": "ns/op\t  58.72 MB/s\t   27760 B/op\t     412 allocs/op",
            "extra": "31839 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_Match (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 37452,
            "unit": "ns/op",
            "extra": "31839 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_Match (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 58.72,
            "unit": "MB/s",
            "extra": "31839 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_Match (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 27760,
            "unit": "B/op",
            "extra": "31839 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_Match (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 412,
            "unit": "allocs/op",
            "extra": "31839 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_Match (github.com/alex60217101990/terse/internal/detect)",
            "value": 37270,
            "unit": "ns/op\t  59.00 MB/s\t   27760 B/op\t     412 allocs/op",
            "extra": "32233 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_Match (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 37270,
            "unit": "ns/op",
            "extra": "32233 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_Match (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 59,
            "unit": "MB/s",
            "extra": "32233 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_Match (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 27760,
            "unit": "B/op",
            "extra": "32233 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_Match (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 412,
            "unit": "allocs/op",
            "extra": "32233 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_Match (github.com/alex60217101990/terse/internal/detect)",
            "value": 37480,
            "unit": "ns/op\t  58.67 MB/s\t   27760 B/op\t     412 allocs/op",
            "extra": "31762 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_Match (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 37480,
            "unit": "ns/op",
            "extra": "31762 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_Match (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 58.67,
            "unit": "MB/s",
            "extra": "31762 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_Match (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 27760,
            "unit": "B/op",
            "extra": "31762 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_Match (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 412,
            "unit": "allocs/op",
            "extra": "31762 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_Match (github.com/alex60217101990/terse/internal/detect)",
            "value": 37588,
            "unit": "ns/op\t  58.50 MB/s\t   27760 B/op\t     412 allocs/op",
            "extra": "31918 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_Match (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 37588,
            "unit": "ns/op",
            "extra": "31918 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_Match (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 58.5,
            "unit": "MB/s",
            "extra": "31918 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_Match (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 27760,
            "unit": "B/op",
            "extra": "31918 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_Match (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 412,
            "unit": "allocs/op",
            "extra": "31918 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_Match (github.com/alex60217101990/terse/internal/detect)",
            "value": 37652,
            "unit": "ns/op\t  58.40 MB/s\t   27760 B/op\t     412 allocs/op",
            "extra": "31934 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_Match (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 37652,
            "unit": "ns/op",
            "extra": "31934 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_Match (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 58.4,
            "unit": "MB/s",
            "extra": "31934 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_Match (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 27760,
            "unit": "B/op",
            "extra": "31934 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_Match (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 412,
            "unit": "allocs/op",
            "extra": "31934 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_Match (github.com/alex60217101990/terse/internal/detect)",
            "value": 38169,
            "unit": "ns/op\t  57.61 MB/s\t   27760 B/op\t     412 allocs/op",
            "extra": "32012 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_Match (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 38169,
            "unit": "ns/op",
            "extra": "32012 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_Match (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 57.61,
            "unit": "MB/s",
            "extra": "32012 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_Match (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 27760,
            "unit": "B/op",
            "extra": "32012 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_Match (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 412,
            "unit": "allocs/op",
            "extra": "32012 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_NonMatch (github.com/alex60217101990/terse/internal/detect)",
            "value": 2.819,
            "unit": "ns/op\t780385.55 MB/s\t       0 B/op\t       0 allocs/op",
            "extra": "426011128 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_NonMatch (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 2.819,
            "unit": "ns/op",
            "extra": "426011128 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_NonMatch (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 780385.55,
            "unit": "MB/s",
            "extra": "426011128 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_NonMatch (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "426011128 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_NonMatch (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "426011128 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_NonMatch (github.com/alex60217101990/terse/internal/detect)",
            "value": 2.821,
            "unit": "ns/op\t779867.38 MB/s\t       0 B/op\t       0 allocs/op",
            "extra": "425458466 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_NonMatch (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 2.821,
            "unit": "ns/op",
            "extra": "425458466 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_NonMatch (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 779867.38,
            "unit": "MB/s",
            "extra": "425458466 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_NonMatch (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "425458466 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_NonMatch (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "425458466 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_NonMatch (github.com/alex60217101990/terse/internal/detect)",
            "value": 2.823,
            "unit": "ns/op\t779426.26 MB/s\t       0 B/op\t       0 allocs/op",
            "extra": "422474607 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_NonMatch (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 2.823,
            "unit": "ns/op",
            "extra": "422474607 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_NonMatch (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 779426.26,
            "unit": "MB/s",
            "extra": "422474607 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_NonMatch (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "422474607 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_NonMatch (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "422474607 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_NonMatch (github.com/alex60217101990/terse/internal/detect)",
            "value": 2.819,
            "unit": "ns/op\t780438.97 MB/s\t       0 B/op\t       0 allocs/op",
            "extra": "424231112 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_NonMatch (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 2.819,
            "unit": "ns/op",
            "extra": "424231112 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_NonMatch (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 780438.97,
            "unit": "MB/s",
            "extra": "424231112 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_NonMatch (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "424231112 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_NonMatch (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "424231112 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_NonMatch (github.com/alex60217101990/terse/internal/detect)",
            "value": 2.819,
            "unit": "ns/op\t780556.55 MB/s\t       0 B/op\t       0 allocs/op",
            "extra": "426217563 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_NonMatch (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 2.819,
            "unit": "ns/op",
            "extra": "426217563 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_NonMatch (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 780556.55,
            "unit": "MB/s",
            "extra": "426217563 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_NonMatch (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "426217563 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_NonMatch (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "426217563 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_NonMatch (github.com/alex60217101990/terse/internal/detect)",
            "value": 2.819,
            "unit": "ns/op\t780391.74 MB/s\t       0 B/op\t       0 allocs/op",
            "extra": "425669834 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_NonMatch (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 2.819,
            "unit": "ns/op",
            "extra": "425669834 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_NonMatch (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 780391.74,
            "unit": "MB/s",
            "extra": "425669834 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_NonMatch (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "425669834 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeJSONObject_NonMatch (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "425669834 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Clean (github.com/alex60217101990/terse/internal/detect)",
            "value": 6009,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "199738 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Clean (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 6009,
            "unit": "ns/op",
            "extra": "199738 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Clean (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "199738 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Clean (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "199738 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Clean (github.com/alex60217101990/terse/internal/detect)",
            "value": 6008,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "199171 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Clean (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 6008,
            "unit": "ns/op",
            "extra": "199171 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Clean (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "199171 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Clean (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "199171 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Clean (github.com/alex60217101990/terse/internal/detect)",
            "value": 6056,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "199542 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Clean (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 6056,
            "unit": "ns/op",
            "extra": "199542 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Clean (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "199542 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Clean (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "199542 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Clean (github.com/alex60217101990/terse/internal/detect)",
            "value": 6076,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "197648 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Clean (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 6076,
            "unit": "ns/op",
            "extra": "197648 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Clean (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "197648 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Clean (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "197648 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Clean (github.com/alex60217101990/terse/internal/detect)",
            "value": 6009,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "199510 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Clean (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 6009,
            "unit": "ns/op",
            "extra": "199510 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Clean (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "199510 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Clean (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "199510 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Clean (github.com/alex60217101990/terse/internal/detect)",
            "value": 6010,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "199706 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Clean (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 6010,
            "unit": "ns/op",
            "extra": "199706 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Clean (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "199706 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Clean (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "199706 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Dirty (github.com/alex60217101990/terse/internal/detect)",
            "value": 5657,
            "unit": "ns/op\t    2688 B/op\t       1 allocs/op",
            "extra": "203548 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Dirty (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 5657,
            "unit": "ns/op",
            "extra": "203548 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Dirty (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 2688,
            "unit": "B/op",
            "extra": "203548 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Dirty (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "203548 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Dirty (github.com/alex60217101990/terse/internal/detect)",
            "value": 5670,
            "unit": "ns/op\t    2688 B/op\t       1 allocs/op",
            "extra": "209025 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Dirty (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 5670,
            "unit": "ns/op",
            "extra": "209025 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Dirty (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 2688,
            "unit": "B/op",
            "extra": "209025 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Dirty (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "209025 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Dirty (github.com/alex60217101990/terse/internal/detect)",
            "value": 5623,
            "unit": "ns/op\t    2688 B/op\t       1 allocs/op",
            "extra": "211158 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Dirty (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 5623,
            "unit": "ns/op",
            "extra": "211158 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Dirty (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 2688,
            "unit": "B/op",
            "extra": "211158 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Dirty (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "211158 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Dirty (github.com/alex60217101990/terse/internal/detect)",
            "value": 5739,
            "unit": "ns/op\t    2688 B/op\t       1 allocs/op",
            "extra": "205124 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Dirty (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 5739,
            "unit": "ns/op",
            "extra": "205124 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Dirty (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 2688,
            "unit": "B/op",
            "extra": "205124 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Dirty (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "205124 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Dirty (github.com/alex60217101990/terse/internal/detect)",
            "value": 5667,
            "unit": "ns/op\t    2688 B/op\t       1 allocs/op",
            "extra": "209050 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Dirty (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 5667,
            "unit": "ns/op",
            "extra": "209050 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Dirty (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 2688,
            "unit": "B/op",
            "extra": "209050 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Dirty (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "209050 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Dirty (github.com/alex60217101990/terse/internal/detect)",
            "value": 5777,
            "unit": "ns/op\t    2688 B/op\t       1 allocs/op",
            "extra": "214933 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Dirty (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 5777,
            "unit": "ns/op",
            "extra": "214933 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Dirty (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 2688,
            "unit": "B/op",
            "extra": "214933 times\n4 procs"
          },
          {
            "name": "BenchmarkStripNoise_Dirty (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "214933 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_Match (github.com/alex60217101990/terse/internal/detect)",
            "value": 7682,
            "unit": "ns/op\t 380.12 MB/s\t    1664 B/op\t       2 allocs/op",
            "extra": "155517 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_Match (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 7682,
            "unit": "ns/op",
            "extra": "155517 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_Match (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 380.12,
            "unit": "MB/s",
            "extra": "155517 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_Match (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 1664,
            "unit": "B/op",
            "extra": "155517 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_Match (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "155517 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_Match (github.com/alex60217101990/terse/internal/detect)",
            "value": 7656,
            "unit": "ns/op\t 381.40 MB/s\t    1664 B/op\t       2 allocs/op",
            "extra": "155883 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_Match (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 7656,
            "unit": "ns/op",
            "extra": "155883 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_Match (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 381.4,
            "unit": "MB/s",
            "extra": "155883 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_Match (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 1664,
            "unit": "B/op",
            "extra": "155883 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_Match (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "155883 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_Match (github.com/alex60217101990/terse/internal/detect)",
            "value": 7656,
            "unit": "ns/op\t 381.40 MB/s\t    1664 B/op\t       2 allocs/op",
            "extra": "153134 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_Match (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 7656,
            "unit": "ns/op",
            "extra": "153134 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_Match (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 381.4,
            "unit": "MB/s",
            "extra": "153134 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_Match (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 1664,
            "unit": "B/op",
            "extra": "153134 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_Match (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "153134 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_Match (github.com/alex60217101990/terse/internal/detect)",
            "value": 7684,
            "unit": "ns/op\t 380.02 MB/s\t    1664 B/op\t       2 allocs/op",
            "extra": "153630 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_Match (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 7684,
            "unit": "ns/op",
            "extra": "153630 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_Match (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 380.02,
            "unit": "MB/s",
            "extra": "153630 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_Match (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 1664,
            "unit": "B/op",
            "extra": "153630 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_Match (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "153630 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_Match (github.com/alex60217101990/terse/internal/detect)",
            "value": 7680,
            "unit": "ns/op\t 380.22 MB/s\t    1664 B/op\t       2 allocs/op",
            "extra": "154736 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_Match (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 7680,
            "unit": "ns/op",
            "extra": "154736 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_Match (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 380.22,
            "unit": "MB/s",
            "extra": "154736 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_Match (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 1664,
            "unit": "B/op",
            "extra": "154736 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_Match (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "154736 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_Match (github.com/alex60217101990/terse/internal/detect)",
            "value": 7687,
            "unit": "ns/op\t 379.85 MB/s\t    1664 B/op\t       2 allocs/op",
            "extra": "153442 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_Match (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 7687,
            "unit": "ns/op",
            "extra": "153442 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_Match (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 379.85,
            "unit": "MB/s",
            "extra": "153442 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_Match (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 1664,
            "unit": "B/op",
            "extra": "153442 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_Match (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "153442 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_NonMatch (github.com/alex60217101990/terse/internal/detect)",
            "value": 432.1,
            "unit": "ns/op\t2499.36 MB/s\t       0 B/op\t       0 allocs/op",
            "extra": "2776274 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_NonMatch (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 432.1,
            "unit": "ns/op",
            "extra": "2776274 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_NonMatch (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 2499.36,
            "unit": "MB/s",
            "extra": "2776274 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_NonMatch (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "2776274 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_NonMatch (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "2776274 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_NonMatch (github.com/alex60217101990/terse/internal/detect)",
            "value": 432.9,
            "unit": "ns/op\t2494.83 MB/s\t       0 B/op\t       0 allocs/op",
            "extra": "2772164 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_NonMatch (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 432.9,
            "unit": "ns/op",
            "extra": "2772164 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_NonMatch (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 2494.83,
            "unit": "MB/s",
            "extra": "2772164 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_NonMatch (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "2772164 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_NonMatch (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "2772164 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_NonMatch (github.com/alex60217101990/terse/internal/detect)",
            "value": 439.1,
            "unit": "ns/op\t2459.59 MB/s\t       0 B/op\t       0 allocs/op",
            "extra": "2734024 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_NonMatch (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 439.1,
            "unit": "ns/op",
            "extra": "2734024 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_NonMatch (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 2459.59,
            "unit": "MB/s",
            "extra": "2734024 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_NonMatch (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "2734024 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_NonMatch (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "2734024 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_NonMatch (github.com/alex60217101990/terse/internal/detect)",
            "value": 433,
            "unit": "ns/op\t2493.95 MB/s\t       0 B/op\t       0 allocs/op",
            "extra": "2767969 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_NonMatch (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 433,
            "unit": "ns/op",
            "extra": "2767969 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_NonMatch (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 2493.95,
            "unit": "MB/s",
            "extra": "2767969 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_NonMatch (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "2767969 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_NonMatch (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "2767969 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_NonMatch (github.com/alex60217101990/terse/internal/detect)",
            "value": 438.2,
            "unit": "ns/op\t2464.88 MB/s\t       0 B/op\t       0 allocs/op",
            "extra": "2734780 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_NonMatch (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 438.2,
            "unit": "ns/op",
            "extra": "2734780 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_NonMatch (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 2464.88,
            "unit": "MB/s",
            "extra": "2734780 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_NonMatch (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "2734780 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_NonMatch (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "2734780 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_NonMatch (github.com/alex60217101990/terse/internal/detect)",
            "value": 431.9,
            "unit": "ns/op\t2500.52 MB/s\t       0 B/op\t       0 allocs/op",
            "extra": "2777875 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_NonMatch (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 431.9,
            "unit": "ns/op",
            "extra": "2777875 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_NonMatch (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 2500.52,
            "unit": "MB/s",
            "extra": "2777875 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_NonMatch (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "2777875 times\n4 procs"
          },
          {
            "name": "BenchmarkFoldPathPrefix_NonMatch (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "2777875 times\n4 procs"
          },
          {
            "name": "BenchmarkSqueezeOutput (github.com/alex60217101990/terse/internal/detect)",
            "value": 151389,
            "unit": "ns/op\t   50831 B/op\t      18 allocs/op",
            "extra": "7003 times\n4 procs"
          },
          {
            "name": "BenchmarkSqueezeOutput (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 151389,
            "unit": "ns/op",
            "extra": "7003 times\n4 procs"
          },
          {
            "name": "BenchmarkSqueezeOutput (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 50831,
            "unit": "B/op",
            "extra": "7003 times\n4 procs"
          },
          {
            "name": "BenchmarkSqueezeOutput (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "7003 times\n4 procs"
          },
          {
            "name": "BenchmarkSqueezeOutput (github.com/alex60217101990/terse/internal/detect)",
            "value": 151063,
            "unit": "ns/op\t   50793 B/op\t      18 allocs/op",
            "extra": "7440 times\n4 procs"
          },
          {
            "name": "BenchmarkSqueezeOutput (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 151063,
            "unit": "ns/op",
            "extra": "7440 times\n4 procs"
          },
          {
            "name": "BenchmarkSqueezeOutput (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 50793,
            "unit": "B/op",
            "extra": "7440 times\n4 procs"
          },
          {
            "name": "BenchmarkSqueezeOutput (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "7440 times\n4 procs"
          },
          {
            "name": "BenchmarkSqueezeOutput (github.com/alex60217101990/terse/internal/detect)",
            "value": 149856,
            "unit": "ns/op\t   50812 B/op\t      18 allocs/op",
            "extra": "7849 times\n4 procs"
          },
          {
            "name": "BenchmarkSqueezeOutput (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 149856,
            "unit": "ns/op",
            "extra": "7849 times\n4 procs"
          },
          {
            "name": "BenchmarkSqueezeOutput (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 50812,
            "unit": "B/op",
            "extra": "7849 times\n4 procs"
          },
          {
            "name": "BenchmarkSqueezeOutput (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "7849 times\n4 procs"
          },
          {
            "name": "BenchmarkSqueezeOutput (github.com/alex60217101990/terse/internal/detect)",
            "value": 153148,
            "unit": "ns/op\t   50810 B/op\t      18 allocs/op",
            "extra": "7317 times\n4 procs"
          },
          {
            "name": "BenchmarkSqueezeOutput (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 153148,
            "unit": "ns/op",
            "extra": "7317 times\n4 procs"
          },
          {
            "name": "BenchmarkSqueezeOutput (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 50810,
            "unit": "B/op",
            "extra": "7317 times\n4 procs"
          },
          {
            "name": "BenchmarkSqueezeOutput (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "7317 times\n4 procs"
          },
          {
            "name": "BenchmarkSqueezeOutput (github.com/alex60217101990/terse/internal/detect)",
            "value": 150703,
            "unit": "ns/op\t   50793 B/op\t      18 allocs/op",
            "extra": "7825 times\n4 procs"
          },
          {
            "name": "BenchmarkSqueezeOutput (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 150703,
            "unit": "ns/op",
            "extra": "7825 times\n4 procs"
          },
          {
            "name": "BenchmarkSqueezeOutput (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 50793,
            "unit": "B/op",
            "extra": "7825 times\n4 procs"
          },
          {
            "name": "BenchmarkSqueezeOutput (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "7825 times\n4 procs"
          },
          {
            "name": "BenchmarkSqueezeOutput (github.com/alex60217101990/terse/internal/detect)",
            "value": 149666,
            "unit": "ns/op\t   50803 B/op\t      18 allocs/op",
            "extra": "7810 times\n4 procs"
          },
          {
            "name": "BenchmarkSqueezeOutput (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 149666,
            "unit": "ns/op",
            "extra": "7810 times\n4 procs"
          },
          {
            "name": "BenchmarkSqueezeOutput (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 50803,
            "unit": "B/op",
            "extra": "7810 times\n4 procs"
          },
          {
            "name": "BenchmarkSqueezeOutput (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "7810 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_Python (github.com/alex60217101990/terse/internal/detect)",
            "value": 1268,
            "unit": "ns/op\t1909.06 MB/s\t    3040 B/op\t      10 allocs/op",
            "extra": "960291 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_Python (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 1268,
            "unit": "ns/op",
            "extra": "960291 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_Python (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 1909.06,
            "unit": "MB/s",
            "extra": "960291 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_Python (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 3040,
            "unit": "B/op",
            "extra": "960291 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_Python (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 10,
            "unit": "allocs/op",
            "extra": "960291 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_Python (github.com/alex60217101990/terse/internal/detect)",
            "value": 1283,
            "unit": "ns/op\t1885.64 MB/s\t    3040 B/op\t      10 allocs/op",
            "extra": "931860 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_Python (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 1283,
            "unit": "ns/op",
            "extra": "931860 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_Python (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 1885.64,
            "unit": "MB/s",
            "extra": "931860 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_Python (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 3040,
            "unit": "B/op",
            "extra": "931860 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_Python (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 10,
            "unit": "allocs/op",
            "extra": "931860 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_Python (github.com/alex60217101990/terse/internal/detect)",
            "value": 1256,
            "unit": "ns/op\t1926.89 MB/s\t    3040 B/op\t      10 allocs/op",
            "extra": "890928 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_Python (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 1256,
            "unit": "ns/op",
            "extra": "890928 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_Python (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 1926.89,
            "unit": "MB/s",
            "extra": "890928 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_Python (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 3040,
            "unit": "B/op",
            "extra": "890928 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_Python (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 10,
            "unit": "allocs/op",
            "extra": "890928 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_Python (github.com/alex60217101990/terse/internal/detect)",
            "value": 1267,
            "unit": "ns/op\t1910.58 MB/s\t    3040 B/op\t      10 allocs/op",
            "extra": "972772 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_Python (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 1267,
            "unit": "ns/op",
            "extra": "972772 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_Python (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 1910.58,
            "unit": "MB/s",
            "extra": "972772 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_Python (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 3040,
            "unit": "B/op",
            "extra": "972772 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_Python (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 10,
            "unit": "allocs/op",
            "extra": "972772 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_Python (github.com/alex60217101990/terse/internal/detect)",
            "value": 1291,
            "unit": "ns/op\t1874.43 MB/s\t    3040 B/op\t      10 allocs/op",
            "extra": "894700 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_Python (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 1291,
            "unit": "ns/op",
            "extra": "894700 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_Python (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 1874.43,
            "unit": "MB/s",
            "extra": "894700 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_Python (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 3040,
            "unit": "B/op",
            "extra": "894700 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_Python (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 10,
            "unit": "allocs/op",
            "extra": "894700 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_Python (github.com/alex60217101990/terse/internal/detect)",
            "value": 1298,
            "unit": "ns/op\t1864.20 MB/s\t    3040 B/op\t      10 allocs/op",
            "extra": "933318 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_Python (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 1298,
            "unit": "ns/op",
            "extra": "933318 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_Python (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 1864.2,
            "unit": "MB/s",
            "extra": "933318 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_Python (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 3040,
            "unit": "B/op",
            "extra": "933318 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_Python (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 10,
            "unit": "allocs/op",
            "extra": "933318 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_GoPanic (github.com/alex60217101990/terse/internal/detect)",
            "value": 1415,
            "unit": "ns/op\t1661.51 MB/s\t    3104 B/op\t       8 allocs/op",
            "extra": "855684 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_GoPanic (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 1415,
            "unit": "ns/op",
            "extra": "855684 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_GoPanic (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 1661.51,
            "unit": "MB/s",
            "extra": "855684 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_GoPanic (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 3104,
            "unit": "B/op",
            "extra": "855684 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_GoPanic (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "855684 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_GoPanic (github.com/alex60217101990/terse/internal/detect)",
            "value": 1433,
            "unit": "ns/op\t1640.58 MB/s\t    3104 B/op\t       8 allocs/op",
            "extra": "787576 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_GoPanic (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 1433,
            "unit": "ns/op",
            "extra": "787576 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_GoPanic (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 1640.58,
            "unit": "MB/s",
            "extra": "787576 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_GoPanic (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 3104,
            "unit": "B/op",
            "extra": "787576 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_GoPanic (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "787576 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_GoPanic (github.com/alex60217101990/terse/internal/detect)",
            "value": 1418,
            "unit": "ns/op\t1657.96 MB/s\t    3104 B/op\t       8 allocs/op",
            "extra": "832560 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_GoPanic (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 1418,
            "unit": "ns/op",
            "extra": "832560 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_GoPanic (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 1657.96,
            "unit": "MB/s",
            "extra": "832560 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_GoPanic (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 3104,
            "unit": "B/op",
            "extra": "832560 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_GoPanic (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "832560 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_GoPanic (github.com/alex60217101990/terse/internal/detect)",
            "value": 1433,
            "unit": "ns/op\t1640.76 MB/s\t    3104 B/op\t       8 allocs/op",
            "extra": "861056 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_GoPanic (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 1433,
            "unit": "ns/op",
            "extra": "861056 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_GoPanic (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 1640.76,
            "unit": "MB/s",
            "extra": "861056 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_GoPanic (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 3104,
            "unit": "B/op",
            "extra": "861056 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_GoPanic (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "861056 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_GoPanic (github.com/alex60217101990/terse/internal/detect)",
            "value": 1416,
            "unit": "ns/op\t1660.34 MB/s\t    3104 B/op\t       8 allocs/op",
            "extra": "807261 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_GoPanic (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 1416,
            "unit": "ns/op",
            "extra": "807261 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_GoPanic (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 1660.34,
            "unit": "MB/s",
            "extra": "807261 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_GoPanic (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 3104,
            "unit": "B/op",
            "extra": "807261 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_GoPanic (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "807261 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_GoPanic (github.com/alex60217101990/terse/internal/detect)",
            "value": 1422,
            "unit": "ns/op\t1653.74 MB/s\t    3104 B/op\t       8 allocs/op",
            "extra": "838416 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_GoPanic (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 1422,
            "unit": "ns/op",
            "extra": "838416 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_GoPanic (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 1653.74,
            "unit": "MB/s",
            "extra": "838416 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_GoPanic (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 3104,
            "unit": "B/op",
            "extra": "838416 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_GoPanic (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "838416 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_NonMatch (github.com/alex60217101990/terse/internal/detect)",
            "value": 272.3,
            "unit": "ns/op\t19830.05 MB/s\t       0 B/op\t       0 allocs/op",
            "extra": "4405722 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_NonMatch (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 272.3,
            "unit": "ns/op",
            "extra": "4405722 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_NonMatch (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 19830.05,
            "unit": "MB/s",
            "extra": "4405722 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_NonMatch (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "4405722 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_NonMatch (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "4405722 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_NonMatch (github.com/alex60217101990/terse/internal/detect)",
            "value": 272.4,
            "unit": "ns/op\t19824.78 MB/s\t       0 B/op\t       0 allocs/op",
            "extra": "4403445 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_NonMatch (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 272.4,
            "unit": "ns/op",
            "extra": "4403445 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_NonMatch (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 19824.78,
            "unit": "MB/s",
            "extra": "4403445 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_NonMatch (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "4403445 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_NonMatch (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "4403445 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_NonMatch (github.com/alex60217101990/terse/internal/detect)",
            "value": 272.4,
            "unit": "ns/op\t19827.13 MB/s\t       0 B/op\t       0 allocs/op",
            "extra": "4404222 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_NonMatch (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 272.4,
            "unit": "ns/op",
            "extra": "4404222 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_NonMatch (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 19827.13,
            "unit": "MB/s",
            "extra": "4404222 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_NonMatch (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "4404222 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_NonMatch (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "4404222 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_NonMatch (github.com/alex60217101990/terse/internal/detect)",
            "value": 272.5,
            "unit": "ns/op\t19819.34 MB/s\t       0 B/op\t       0 allocs/op",
            "extra": "4404090 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_NonMatch (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 272.5,
            "unit": "ns/op",
            "extra": "4404090 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_NonMatch (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 19819.34,
            "unit": "MB/s",
            "extra": "4404090 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_NonMatch (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "4404090 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_NonMatch (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "4404090 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_NonMatch (github.com/alex60217101990/terse/internal/detect)",
            "value": 272.2,
            "unit": "ns/op\t19835.77 MB/s\t       0 B/op\t       0 allocs/op",
            "extra": "4402701 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_NonMatch (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 272.2,
            "unit": "ns/op",
            "extra": "4402701 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_NonMatch (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 19835.77,
            "unit": "MB/s",
            "extra": "4402701 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_NonMatch (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "4402701 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_NonMatch (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "4402701 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_NonMatch (github.com/alex60217101990/terse/internal/detect)",
            "value": 272.8,
            "unit": "ns/op\t19793.05 MB/s\t       0 B/op\t       0 allocs/op",
            "extra": "4402461 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_NonMatch (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 272.8,
            "unit": "ns/op",
            "extra": "4402461 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_NonMatch (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 19793.05,
            "unit": "MB/s",
            "extra": "4402461 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_NonMatch (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "4402461 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeStackTrace_NonMatch (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "4402461 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable (github.com/alex60217101990/terse/internal/detect)",
            "value": 1083,
            "unit": "ns/op\t2628.98 MB/s\t    1056 B/op\t       4 allocs/op",
            "extra": "993225 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 1083,
            "unit": "ns/op",
            "extra": "993225 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 2628.98,
            "unit": "MB/s",
            "extra": "993225 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 1056,
            "unit": "B/op",
            "extra": "993225 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "993225 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable (github.com/alex60217101990/terse/internal/detect)",
            "value": 1095,
            "unit": "ns/op\t2599.85 MB/s\t    1056 B/op\t       4 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 1095,
            "unit": "ns/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 2599.85,
            "unit": "MB/s",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 1056,
            "unit": "B/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable (github.com/alex60217101990/terse/internal/detect)",
            "value": 1081,
            "unit": "ns/op\t2633.51 MB/s\t    1056 B/op\t       4 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 1081,
            "unit": "ns/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 2633.51,
            "unit": "MB/s",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 1056,
            "unit": "B/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable (github.com/alex60217101990/terse/internal/detect)",
            "value": 1104,
            "unit": "ns/op\t2576.81 MB/s\t    1056 B/op\t       4 allocs/op",
            "extra": "1083626 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 1104,
            "unit": "ns/op",
            "extra": "1083626 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 2576.81,
            "unit": "MB/s",
            "extra": "1083626 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 1056,
            "unit": "B/op",
            "extra": "1083626 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "1083626 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable (github.com/alex60217101990/terse/internal/detect)",
            "value": 1072,
            "unit": "ns/op\t2654.10 MB/s\t    1056 B/op\t       4 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 1072,
            "unit": "ns/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 2654.1,
            "unit": "MB/s",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 1056,
            "unit": "B/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable (github.com/alex60217101990/terse/internal/detect)",
            "value": 1089,
            "unit": "ns/op\t2612.51 MB/s\t    1056 B/op\t       4 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 1089,
            "unit": "ns/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 2612.51,
            "unit": "MB/s",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 1056,
            "unit": "B/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable_NonMatch (github.com/alex60217101990/terse/internal/detect)",
            "value": 29.43,
            "unit": "ns/op\t57088.02 MB/s\t       0 B/op\t       0 allocs/op",
            "extra": "40649619 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable_NonMatch (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 29.43,
            "unit": "ns/op",
            "extra": "40649619 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable_NonMatch (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 57088.02,
            "unit": "MB/s",
            "extra": "40649619 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable_NonMatch (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "40649619 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable_NonMatch (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "40649619 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable_NonMatch (github.com/alex60217101990/terse/internal/detect)",
            "value": 29.55,
            "unit": "ns/op\t56856.61 MB/s\t       0 B/op\t       0 allocs/op",
            "extra": "40690743 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable_NonMatch (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 29.55,
            "unit": "ns/op",
            "extra": "40690743 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable_NonMatch (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 56856.61,
            "unit": "MB/s",
            "extra": "40690743 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable_NonMatch (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "40690743 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable_NonMatch (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "40690743 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable_NonMatch (github.com/alex60217101990/terse/internal/detect)",
            "value": 29.44,
            "unit": "ns/op\t57060.41 MB/s\t       0 B/op\t       0 allocs/op",
            "extra": "40614111 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable_NonMatch (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 29.44,
            "unit": "ns/op",
            "extra": "40614111 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable_NonMatch (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 57060.41,
            "unit": "MB/s",
            "extra": "40614111 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable_NonMatch (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "40614111 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable_NonMatch (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "40614111 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable_NonMatch (github.com/alex60217101990/terse/internal/detect)",
            "value": 29.44,
            "unit": "ns/op\t57073.34 MB/s\t       0 B/op\t       0 allocs/op",
            "extra": "40735586 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable_NonMatch (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 29.44,
            "unit": "ns/op",
            "extra": "40735586 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable_NonMatch (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 57073.34,
            "unit": "MB/s",
            "extra": "40735586 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable_NonMatch (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "40735586 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable_NonMatch (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "40735586 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable_NonMatch (github.com/alex60217101990/terse/internal/detect)",
            "value": 29.5,
            "unit": "ns/op\t56950.18 MB/s\t       0 B/op\t       0 allocs/op",
            "extra": "40755177 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable_NonMatch (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 29.5,
            "unit": "ns/op",
            "extra": "40755177 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable_NonMatch (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 56950.18,
            "unit": "MB/s",
            "extra": "40755177 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable_NonMatch (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "40755177 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable_NonMatch (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "40755177 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable_NonMatch (github.com/alex60217101990/terse/internal/detect)",
            "value": 29.47,
            "unit": "ns/op\t57001.57 MB/s\t       0 B/op\t       0 allocs/op",
            "extra": "40686660 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable_NonMatch (github.com/alex60217101990/terse/internal/detect) - ns/op",
            "value": 29.47,
            "unit": "ns/op",
            "extra": "40686660 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable_NonMatch (github.com/alex60217101990/terse/internal/detect) - MB/s",
            "value": 57001.57,
            "unit": "MB/s",
            "extra": "40686660 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable_NonMatch (github.com/alex60217101990/terse/internal/detect) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "40686660 times\n4 procs"
          },
          {
            "name": "BenchmarkSummarizeTable_NonMatch (github.com/alex60217101990/terse/internal/detect) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "40686660 times\n4 procs"
          },
          {
            "name": "BenchmarkSliceLines (github.com/alex60217101990/terse/internal/hook)",
            "value": 5191,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "229950 times\n4 procs"
          },
          {
            "name": "BenchmarkSliceLines (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 5191,
            "unit": "ns/op",
            "extra": "229950 times\n4 procs"
          },
          {
            "name": "BenchmarkSliceLines (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "229950 times\n4 procs"
          },
          {
            "name": "BenchmarkSliceLines (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "229950 times\n4 procs"
          },
          {
            "name": "BenchmarkSliceLines (github.com/alex60217101990/terse/internal/hook)",
            "value": 5186,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "231339 times\n4 procs"
          },
          {
            "name": "BenchmarkSliceLines (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 5186,
            "unit": "ns/op",
            "extra": "231339 times\n4 procs"
          },
          {
            "name": "BenchmarkSliceLines (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "231339 times\n4 procs"
          },
          {
            "name": "BenchmarkSliceLines (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "231339 times\n4 procs"
          },
          {
            "name": "BenchmarkSliceLines (github.com/alex60217101990/terse/internal/hook)",
            "value": 5186,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "231136 times\n4 procs"
          },
          {
            "name": "BenchmarkSliceLines (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 5186,
            "unit": "ns/op",
            "extra": "231136 times\n4 procs"
          },
          {
            "name": "BenchmarkSliceLines (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "231136 times\n4 procs"
          },
          {
            "name": "BenchmarkSliceLines (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "231136 times\n4 procs"
          },
          {
            "name": "BenchmarkSliceLines (github.com/alex60217101990/terse/internal/hook)",
            "value": 5201,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "231380 times\n4 procs"
          },
          {
            "name": "BenchmarkSliceLines (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 5201,
            "unit": "ns/op",
            "extra": "231380 times\n4 procs"
          },
          {
            "name": "BenchmarkSliceLines (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "231380 times\n4 procs"
          },
          {
            "name": "BenchmarkSliceLines (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "231380 times\n4 procs"
          },
          {
            "name": "BenchmarkSliceLines (github.com/alex60217101990/terse/internal/hook)",
            "value": 5182,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "231492 times\n4 procs"
          },
          {
            "name": "BenchmarkSliceLines (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 5182,
            "unit": "ns/op",
            "extra": "231492 times\n4 procs"
          },
          {
            "name": "BenchmarkSliceLines (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "231492 times\n4 procs"
          },
          {
            "name": "BenchmarkSliceLines (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "231492 times\n4 procs"
          },
          {
            "name": "BenchmarkSliceLines (github.com/alex60217101990/terse/internal/hook)",
            "value": 5188,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "231176 times\n4 procs"
          },
          {
            "name": "BenchmarkSliceLines (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 5188,
            "unit": "ns/op",
            "extra": "231176 times\n4 procs"
          },
          {
            "name": "BenchmarkSliceLines (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "231176 times\n4 procs"
          },
          {
            "name": "BenchmarkSliceLines (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "231176 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_Unchanged (github.com/alex60217101990/terse/internal/hook)",
            "value": 312730,
            "unit": "ns/op\t   61864 B/op\t      70 allocs/op",
            "extra": "3865 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_Unchanged (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 312730,
            "unit": "ns/op",
            "extra": "3865 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_Unchanged (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 61864,
            "unit": "B/op",
            "extra": "3865 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_Unchanged (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 70,
            "unit": "allocs/op",
            "extra": "3865 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_Unchanged (github.com/alex60217101990/terse/internal/hook)",
            "value": 312218,
            "unit": "ns/op\t   61935 B/op\t      70 allocs/op",
            "extra": "3886 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_Unchanged (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 312218,
            "unit": "ns/op",
            "extra": "3886 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_Unchanged (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 61935,
            "unit": "B/op",
            "extra": "3886 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_Unchanged (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 70,
            "unit": "allocs/op",
            "extra": "3886 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_Unchanged (github.com/alex60217101990/terse/internal/hook)",
            "value": 309109,
            "unit": "ns/op\t   61949 B/op\t      70 allocs/op",
            "extra": "3964 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_Unchanged (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 309109,
            "unit": "ns/op",
            "extra": "3964 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_Unchanged (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 61949,
            "unit": "B/op",
            "extra": "3964 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_Unchanged (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 70,
            "unit": "allocs/op",
            "extra": "3964 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_Unchanged (github.com/alex60217101990/terse/internal/hook)",
            "value": 307800,
            "unit": "ns/op\t   61928 B/op\t      70 allocs/op",
            "extra": "3962 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_Unchanged (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 307800,
            "unit": "ns/op",
            "extra": "3962 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_Unchanged (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 61928,
            "unit": "B/op",
            "extra": "3962 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_Unchanged (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 70,
            "unit": "allocs/op",
            "extra": "3962 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_Unchanged (github.com/alex60217101990/terse/internal/hook)",
            "value": 309878,
            "unit": "ns/op\t   61971 B/op\t      70 allocs/op",
            "extra": "3940 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_Unchanged (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 309878,
            "unit": "ns/op",
            "extra": "3940 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_Unchanged (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 61971,
            "unit": "B/op",
            "extra": "3940 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_Unchanged (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 70,
            "unit": "allocs/op",
            "extra": "3940 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_Unchanged (github.com/alex60217101990/terse/internal/hook)",
            "value": 309152,
            "unit": "ns/op\t   61931 B/op\t      70 allocs/op",
            "extra": "3938 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_Unchanged (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 309152,
            "unit": "ns/op",
            "extra": "3938 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_Unchanged (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 61931,
            "unit": "B/op",
            "extra": "3938 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_Unchanged (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 70,
            "unit": "allocs/op",
            "extra": "3938 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_FirstRead (github.com/alex60217101990/terse/internal/hook)",
            "value": 281934,
            "unit": "ns/op\t   82341 B/op\t      90 allocs/op",
            "extra": "4243 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_FirstRead (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 281934,
            "unit": "ns/op",
            "extra": "4243 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_FirstRead (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 82341,
            "unit": "B/op",
            "extra": "4243 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_FirstRead (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 90,
            "unit": "allocs/op",
            "extra": "4243 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_FirstRead (github.com/alex60217101990/terse/internal/hook)",
            "value": 278613,
            "unit": "ns/op\t   82492 B/op\t      90 allocs/op",
            "extra": "3960 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_FirstRead (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 278613,
            "unit": "ns/op",
            "extra": "3960 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_FirstRead (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 82492,
            "unit": "B/op",
            "extra": "3960 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_FirstRead (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 90,
            "unit": "allocs/op",
            "extra": "3960 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_FirstRead (github.com/alex60217101990/terse/internal/hook)",
            "value": 285553,
            "unit": "ns/op\t   82495 B/op\t      90 allocs/op",
            "extra": "3696 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_FirstRead (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 285553,
            "unit": "ns/op",
            "extra": "3696 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_FirstRead (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 82495,
            "unit": "B/op",
            "extra": "3696 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_FirstRead (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 90,
            "unit": "allocs/op",
            "extra": "3696 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_FirstRead (github.com/alex60217101990/terse/internal/hook)",
            "value": 278230,
            "unit": "ns/op\t   82307 B/op\t      90 allocs/op",
            "extra": "3987 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_FirstRead (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 278230,
            "unit": "ns/op",
            "extra": "3987 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_FirstRead (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 82307,
            "unit": "B/op",
            "extra": "3987 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_FirstRead (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 90,
            "unit": "allocs/op",
            "extra": "3987 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_FirstRead (github.com/alex60217101990/terse/internal/hook)",
            "value": 279184,
            "unit": "ns/op\t   82417 B/op\t      90 allocs/op",
            "extra": "4178 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_FirstRead (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 279184,
            "unit": "ns/op",
            "extra": "4178 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_FirstRead (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 82417,
            "unit": "B/op",
            "extra": "4178 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_FirstRead (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 90,
            "unit": "allocs/op",
            "extra": "4178 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_FirstRead (github.com/alex60217101990/terse/internal/hook)",
            "value": 278058,
            "unit": "ns/op\t   82560 B/op\t      90 allocs/op",
            "extra": "4238 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_FirstRead (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 278058,
            "unit": "ns/op",
            "extra": "4238 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_FirstRead (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 82560,
            "unit": "B/op",
            "extra": "4238 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_FirstRead (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 90,
            "unit": "allocs/op",
            "extra": "4238 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Unchanged (github.com/alex60217101990/terse/internal/hook)",
            "value": 33360,
            "unit": "ns/op\t    4874 B/op\t      45 allocs/op",
            "extra": "35642 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Unchanged (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 33360,
            "unit": "ns/op",
            "extra": "35642 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Unchanged (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 4874,
            "unit": "B/op",
            "extra": "35642 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Unchanged (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 45,
            "unit": "allocs/op",
            "extra": "35642 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Unchanged (github.com/alex60217101990/terse/internal/hook)",
            "value": 33159,
            "unit": "ns/op\t    4889 B/op\t      45 allocs/op",
            "extra": "36176 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Unchanged (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 33159,
            "unit": "ns/op",
            "extra": "36176 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Unchanged (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 4889,
            "unit": "B/op",
            "extra": "36176 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Unchanged (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 45,
            "unit": "allocs/op",
            "extra": "36176 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Unchanged (github.com/alex60217101990/terse/internal/hook)",
            "value": 33276,
            "unit": "ns/op\t    4889 B/op\t      45 allocs/op",
            "extra": "35841 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Unchanged (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 33276,
            "unit": "ns/op",
            "extra": "35841 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Unchanged (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 4889,
            "unit": "B/op",
            "extra": "35841 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Unchanged (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 45,
            "unit": "allocs/op",
            "extra": "35841 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Unchanged (github.com/alex60217101990/terse/internal/hook)",
            "value": 33284,
            "unit": "ns/op\t    4890 B/op\t      45 allocs/op",
            "extra": "36128 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Unchanged (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 33284,
            "unit": "ns/op",
            "extra": "36128 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Unchanged (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 4890,
            "unit": "B/op",
            "extra": "36128 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Unchanged (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 45,
            "unit": "allocs/op",
            "extra": "36128 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Unchanged (github.com/alex60217101990/terse/internal/hook)",
            "value": 33369,
            "unit": "ns/op\t    4889 B/op\t      45 allocs/op",
            "extra": "35716 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Unchanged (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 33369,
            "unit": "ns/op",
            "extra": "35716 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Unchanged (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 4889,
            "unit": "B/op",
            "extra": "35716 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Unchanged (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 45,
            "unit": "allocs/op",
            "extra": "35716 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Unchanged (github.com/alex60217101990/terse/internal/hook)",
            "value": 33370,
            "unit": "ns/op\t    4889 B/op\t      45 allocs/op",
            "extra": "35623 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Unchanged (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 33370,
            "unit": "ns/op",
            "extra": "35623 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Unchanged (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 4889,
            "unit": "B/op",
            "extra": "35623 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Unchanged (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 45,
            "unit": "allocs/op",
            "extra": "35623 times\n4 procs"
          },
          {
            "name": "BenchmarkBashHook_JSONArray (github.com/alex60217101990/terse/internal/hook)",
            "value": 3507062,
            "unit": "ns/op\t  817854 B/op\t     133 allocs/op",
            "extra": "340 times\n4 procs"
          },
          {
            "name": "BenchmarkBashHook_JSONArray (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 3507062,
            "unit": "ns/op",
            "extra": "340 times\n4 procs"
          },
          {
            "name": "BenchmarkBashHook_JSONArray (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 817854,
            "unit": "B/op",
            "extra": "340 times\n4 procs"
          },
          {
            "name": "BenchmarkBashHook_JSONArray (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 133,
            "unit": "allocs/op",
            "extra": "340 times\n4 procs"
          },
          {
            "name": "BenchmarkBashHook_JSONArray (github.com/alex60217101990/terse/internal/hook)",
            "value": 3489144,
            "unit": "ns/op\t  817848 B/op\t     133 allocs/op",
            "extra": "342 times\n4 procs"
          },
          {
            "name": "BenchmarkBashHook_JSONArray (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 3489144,
            "unit": "ns/op",
            "extra": "342 times\n4 procs"
          },
          {
            "name": "BenchmarkBashHook_JSONArray (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 817848,
            "unit": "B/op",
            "extra": "342 times\n4 procs"
          },
          {
            "name": "BenchmarkBashHook_JSONArray (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 133,
            "unit": "allocs/op",
            "extra": "342 times\n4 procs"
          },
          {
            "name": "BenchmarkBashHook_JSONArray (github.com/alex60217101990/terse/internal/hook)",
            "value": 3485591,
            "unit": "ns/op\t  817901 B/op\t     133 allocs/op",
            "extra": "343 times\n4 procs"
          },
          {
            "name": "BenchmarkBashHook_JSONArray (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 3485591,
            "unit": "ns/op",
            "extra": "343 times\n4 procs"
          },
          {
            "name": "BenchmarkBashHook_JSONArray (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 817901,
            "unit": "B/op",
            "extra": "343 times\n4 procs"
          },
          {
            "name": "BenchmarkBashHook_JSONArray (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 133,
            "unit": "allocs/op",
            "extra": "343 times\n4 procs"
          },
          {
            "name": "BenchmarkBashHook_JSONArray (github.com/alex60217101990/terse/internal/hook)",
            "value": 3472900,
            "unit": "ns/op\t  817813 B/op\t     132 allocs/op",
            "extra": "342 times\n4 procs"
          },
          {
            "name": "BenchmarkBashHook_JSONArray (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 3472900,
            "unit": "ns/op",
            "extra": "342 times\n4 procs"
          },
          {
            "name": "BenchmarkBashHook_JSONArray (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 817813,
            "unit": "B/op",
            "extra": "342 times\n4 procs"
          },
          {
            "name": "BenchmarkBashHook_JSONArray (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 132,
            "unit": "allocs/op",
            "extra": "342 times\n4 procs"
          },
          {
            "name": "BenchmarkBashHook_JSONArray (github.com/alex60217101990/terse/internal/hook)",
            "value": 3505656,
            "unit": "ns/op\t  817881 B/op\t     133 allocs/op",
            "extra": "334 times\n4 procs"
          },
          {
            "name": "BenchmarkBashHook_JSONArray (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 3505656,
            "unit": "ns/op",
            "extra": "334 times\n4 procs"
          },
          {
            "name": "BenchmarkBashHook_JSONArray (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 817881,
            "unit": "B/op",
            "extra": "334 times\n4 procs"
          },
          {
            "name": "BenchmarkBashHook_JSONArray (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 133,
            "unit": "allocs/op",
            "extra": "334 times\n4 procs"
          },
          {
            "name": "BenchmarkBashHook_JSONArray (github.com/alex60217101990/terse/internal/hook)",
            "value": 3481369,
            "unit": "ns/op\t  817876 B/op\t     133 allocs/op",
            "extra": "344 times\n4 procs"
          },
          {
            "name": "BenchmarkBashHook_JSONArray (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 3481369,
            "unit": "ns/op",
            "extra": "344 times\n4 procs"
          },
          {
            "name": "BenchmarkBashHook_JSONArray (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 817876,
            "unit": "B/op",
            "extra": "344 times\n4 procs"
          },
          {
            "name": "BenchmarkBashHook_JSONArray (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 133,
            "unit": "allocs/op",
            "extra": "344 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Allow (github.com/alex60217101990/terse/internal/hook)",
            "value": 22847,
            "unit": "ns/op\t    2880 B/op\t      38 allocs/op",
            "extra": "52323 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Allow (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 22847,
            "unit": "ns/op",
            "extra": "52323 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Allow (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 2880,
            "unit": "B/op",
            "extra": "52323 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Allow (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 38,
            "unit": "allocs/op",
            "extra": "52323 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Allow (github.com/alex60217101990/terse/internal/hook)",
            "value": 22718,
            "unit": "ns/op\t    2880 B/op\t      38 allocs/op",
            "extra": "52410 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Allow (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 22718,
            "unit": "ns/op",
            "extra": "52410 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Allow (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 2880,
            "unit": "B/op",
            "extra": "52410 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Allow (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 38,
            "unit": "allocs/op",
            "extra": "52410 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Allow (github.com/alex60217101990/terse/internal/hook)",
            "value": 22743,
            "unit": "ns/op\t    2880 B/op\t      38 allocs/op",
            "extra": "52887 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Allow (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 22743,
            "unit": "ns/op",
            "extra": "52887 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Allow (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 2880,
            "unit": "B/op",
            "extra": "52887 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Allow (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 38,
            "unit": "allocs/op",
            "extra": "52887 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Allow (github.com/alex60217101990/terse/internal/hook)",
            "value": 22704,
            "unit": "ns/op\t    2880 B/op\t      38 allocs/op",
            "extra": "52502 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Allow (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 22704,
            "unit": "ns/op",
            "extra": "52502 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Allow (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 2880,
            "unit": "B/op",
            "extra": "52502 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Allow (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 38,
            "unit": "allocs/op",
            "extra": "52502 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Allow (github.com/alex60217101990/terse/internal/hook)",
            "value": 22691,
            "unit": "ns/op\t    2880 B/op\t      38 allocs/op",
            "extra": "52275 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Allow (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 22691,
            "unit": "ns/op",
            "extra": "52275 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Allow (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 2880,
            "unit": "B/op",
            "extra": "52275 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Allow (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 38,
            "unit": "allocs/op",
            "extra": "52275 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Allow (github.com/alex60217101990/terse/internal/hook)",
            "value": 22634,
            "unit": "ns/op\t    2880 B/op\t      38 allocs/op",
            "extra": "52551 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Allow (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 22634,
            "unit": "ns/op",
            "extra": "52551 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Allow (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 2880,
            "unit": "B/op",
            "extra": "52551 times\n4 procs"
          },
          {
            "name": "BenchmarkPreToolUse_Allow (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 38,
            "unit": "allocs/op",
            "extra": "52551 times\n4 procs"
          },
          {
            "name": "BenchmarkGlob_Tree (github.com/alex60217101990/terse/internal/hook)",
            "value": 55708,
            "unit": "ns/op\t   12277 B/op\t     152 allocs/op",
            "extra": "21652 times\n4 procs"
          },
          {
            "name": "BenchmarkGlob_Tree (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 55708,
            "unit": "ns/op",
            "extra": "21652 times\n4 procs"
          },
          {
            "name": "BenchmarkGlob_Tree (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 12277,
            "unit": "B/op",
            "extra": "21652 times\n4 procs"
          },
          {
            "name": "BenchmarkGlob_Tree (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 152,
            "unit": "allocs/op",
            "extra": "21652 times\n4 procs"
          },
          {
            "name": "BenchmarkGlob_Tree (github.com/alex60217101990/terse/internal/hook)",
            "value": 55963,
            "unit": "ns/op\t   12277 B/op\t     152 allocs/op",
            "extra": "21470 times\n4 procs"
          },
          {
            "name": "BenchmarkGlob_Tree (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 55963,
            "unit": "ns/op",
            "extra": "21470 times\n4 procs"
          },
          {
            "name": "BenchmarkGlob_Tree (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 12277,
            "unit": "B/op",
            "extra": "21470 times\n4 procs"
          },
          {
            "name": "BenchmarkGlob_Tree (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 152,
            "unit": "allocs/op",
            "extra": "21470 times\n4 procs"
          },
          {
            "name": "BenchmarkGlob_Tree (github.com/alex60217101990/terse/internal/hook)",
            "value": 57122,
            "unit": "ns/op\t   12277 B/op\t     152 allocs/op",
            "extra": "20982 times\n4 procs"
          },
          {
            "name": "BenchmarkGlob_Tree (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 57122,
            "unit": "ns/op",
            "extra": "20982 times\n4 procs"
          },
          {
            "name": "BenchmarkGlob_Tree (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 12277,
            "unit": "B/op",
            "extra": "20982 times\n4 procs"
          },
          {
            "name": "BenchmarkGlob_Tree (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 152,
            "unit": "allocs/op",
            "extra": "20982 times\n4 procs"
          },
          {
            "name": "BenchmarkGlob_Tree (github.com/alex60217101990/terse/internal/hook)",
            "value": 56284,
            "unit": "ns/op\t   12277 B/op\t     152 allocs/op",
            "extra": "21361 times\n4 procs"
          },
          {
            "name": "BenchmarkGlob_Tree (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 56284,
            "unit": "ns/op",
            "extra": "21361 times\n4 procs"
          },
          {
            "name": "BenchmarkGlob_Tree (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 12277,
            "unit": "B/op",
            "extra": "21361 times\n4 procs"
          },
          {
            "name": "BenchmarkGlob_Tree (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 152,
            "unit": "allocs/op",
            "extra": "21361 times\n4 procs"
          },
          {
            "name": "BenchmarkGlob_Tree (github.com/alex60217101990/terse/internal/hook)",
            "value": 56287,
            "unit": "ns/op\t   12277 B/op\t     152 allocs/op",
            "extra": "21415 times\n4 procs"
          },
          {
            "name": "BenchmarkGlob_Tree (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 56287,
            "unit": "ns/op",
            "extra": "21415 times\n4 procs"
          },
          {
            "name": "BenchmarkGlob_Tree (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 12277,
            "unit": "B/op",
            "extra": "21415 times\n4 procs"
          },
          {
            "name": "BenchmarkGlob_Tree (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 152,
            "unit": "allocs/op",
            "extra": "21415 times\n4 procs"
          },
          {
            "name": "BenchmarkGlob_Tree (github.com/alex60217101990/terse/internal/hook)",
            "value": 57264,
            "unit": "ns/op\t   12277 B/op\t     152 allocs/op",
            "extra": "21080 times\n4 procs"
          },
          {
            "name": "BenchmarkGlob_Tree (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 57264,
            "unit": "ns/op",
            "extra": "21080 times\n4 procs"
          },
          {
            "name": "BenchmarkGlob_Tree (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 12277,
            "unit": "B/op",
            "extra": "21080 times\n4 procs"
          },
          {
            "name": "BenchmarkGlob_Tree (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 152,
            "unit": "allocs/op",
            "extra": "21080 times\n4 procs"
          },
          {
            "name": "BenchmarkWrite_Compress (github.com/alex60217101990/terse/internal/hook)",
            "value": 30312,
            "unit": "ns/op\t    5495 B/op\t      37 allocs/op",
            "extra": "37645 times\n4 procs"
          },
          {
            "name": "BenchmarkWrite_Compress (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 30312,
            "unit": "ns/op",
            "extra": "37645 times\n4 procs"
          },
          {
            "name": "BenchmarkWrite_Compress (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 5495,
            "unit": "B/op",
            "extra": "37645 times\n4 procs"
          },
          {
            "name": "BenchmarkWrite_Compress (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 37,
            "unit": "allocs/op",
            "extra": "37645 times\n4 procs"
          },
          {
            "name": "BenchmarkWrite_Compress (github.com/alex60217101990/terse/internal/hook)",
            "value": 29938,
            "unit": "ns/op\t    5496 B/op\t      37 allocs/op",
            "extra": "40150 times\n4 procs"
          },
          {
            "name": "BenchmarkWrite_Compress (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 29938,
            "unit": "ns/op",
            "extra": "40150 times\n4 procs"
          },
          {
            "name": "BenchmarkWrite_Compress (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 5496,
            "unit": "B/op",
            "extra": "40150 times\n4 procs"
          },
          {
            "name": "BenchmarkWrite_Compress (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 37,
            "unit": "allocs/op",
            "extra": "40150 times\n4 procs"
          },
          {
            "name": "BenchmarkWrite_Compress (github.com/alex60217101990/terse/internal/hook)",
            "value": 30265,
            "unit": "ns/op\t    5496 B/op\t      37 allocs/op",
            "extra": "40267 times\n4 procs"
          },
          {
            "name": "BenchmarkWrite_Compress (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 30265,
            "unit": "ns/op",
            "extra": "40267 times\n4 procs"
          },
          {
            "name": "BenchmarkWrite_Compress (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 5496,
            "unit": "B/op",
            "extra": "40267 times\n4 procs"
          },
          {
            "name": "BenchmarkWrite_Compress (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 37,
            "unit": "allocs/op",
            "extra": "40267 times\n4 procs"
          },
          {
            "name": "BenchmarkWrite_Compress (github.com/alex60217101990/terse/internal/hook)",
            "value": 29906,
            "unit": "ns/op\t    5496 B/op\t      37 allocs/op",
            "extra": "39856 times\n4 procs"
          },
          {
            "name": "BenchmarkWrite_Compress (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 29906,
            "unit": "ns/op",
            "extra": "39856 times\n4 procs"
          },
          {
            "name": "BenchmarkWrite_Compress (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 5496,
            "unit": "B/op",
            "extra": "39856 times\n4 procs"
          },
          {
            "name": "BenchmarkWrite_Compress (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 37,
            "unit": "allocs/op",
            "extra": "39856 times\n4 procs"
          },
          {
            "name": "BenchmarkWrite_Compress (github.com/alex60217101990/terse/internal/hook)",
            "value": 29814,
            "unit": "ns/op\t    5496 B/op\t      37 allocs/op",
            "extra": "39880 times\n4 procs"
          },
          {
            "name": "BenchmarkWrite_Compress (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 29814,
            "unit": "ns/op",
            "extra": "39880 times\n4 procs"
          },
          {
            "name": "BenchmarkWrite_Compress (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 5496,
            "unit": "B/op",
            "extra": "39880 times\n4 procs"
          },
          {
            "name": "BenchmarkWrite_Compress (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 37,
            "unit": "allocs/op",
            "extra": "39880 times\n4 procs"
          },
          {
            "name": "BenchmarkWrite_Compress (github.com/alex60217101990/terse/internal/hook)",
            "value": 29589,
            "unit": "ns/op\t    5496 B/op\t      37 allocs/op",
            "extra": "40530 times\n4 procs"
          },
          {
            "name": "BenchmarkWrite_Compress (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 29589,
            "unit": "ns/op",
            "extra": "40530 times\n4 procs"
          },
          {
            "name": "BenchmarkWrite_Compress (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 5496,
            "unit": "B/op",
            "extra": "40530 times\n4 procs"
          },
          {
            "name": "BenchmarkWrite_Compress (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 37,
            "unit": "allocs/op",
            "extra": "40530 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_PreToolIntercept (github.com/alex60217101990/terse/internal/hook)",
            "value": 33885,
            "unit": "ns/op\t    4945 B/op\t      45 allocs/op",
            "extra": "35284 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_PreToolIntercept (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 33885,
            "unit": "ns/op",
            "extra": "35284 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_PreToolIntercept (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 4945,
            "unit": "B/op",
            "extra": "35284 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_PreToolIntercept (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 45,
            "unit": "allocs/op",
            "extra": "35284 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_PreToolIntercept (github.com/alex60217101990/terse/internal/hook)",
            "value": 33490,
            "unit": "ns/op\t    4946 B/op\t      45 allocs/op",
            "extra": "35478 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_PreToolIntercept (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 33490,
            "unit": "ns/op",
            "extra": "35478 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_PreToolIntercept (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 4946,
            "unit": "B/op",
            "extra": "35478 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_PreToolIntercept (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 45,
            "unit": "allocs/op",
            "extra": "35478 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_PreToolIntercept (github.com/alex60217101990/terse/internal/hook)",
            "value": 33500,
            "unit": "ns/op\t    4945 B/op\t      45 allocs/op",
            "extra": "35766 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_PreToolIntercept (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 33500,
            "unit": "ns/op",
            "extra": "35766 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_PreToolIntercept (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 4945,
            "unit": "B/op",
            "extra": "35766 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_PreToolIntercept (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 45,
            "unit": "allocs/op",
            "extra": "35766 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_PreToolIntercept (github.com/alex60217101990/terse/internal/hook)",
            "value": 33508,
            "unit": "ns/op\t    4945 B/op\t      45 allocs/op",
            "extra": "35726 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_PreToolIntercept (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 33508,
            "unit": "ns/op",
            "extra": "35726 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_PreToolIntercept (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 4945,
            "unit": "B/op",
            "extra": "35726 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_PreToolIntercept (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 45,
            "unit": "allocs/op",
            "extra": "35726 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_PreToolIntercept (github.com/alex60217101990/terse/internal/hook)",
            "value": 33313,
            "unit": "ns/op\t    4946 B/op\t      45 allocs/op",
            "extra": "35820 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_PreToolIntercept (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 33313,
            "unit": "ns/op",
            "extra": "35820 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_PreToolIntercept (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 4946,
            "unit": "B/op",
            "extra": "35820 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_PreToolIntercept (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 45,
            "unit": "allocs/op",
            "extra": "35820 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_PreToolIntercept (github.com/alex60217101990/terse/internal/hook)",
            "value": 33427,
            "unit": "ns/op\t    4946 B/op\t      45 allocs/op",
            "extra": "35838 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_PreToolIntercept (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 33427,
            "unit": "ns/op",
            "extra": "35838 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_PreToolIntercept (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 4946,
            "unit": "B/op",
            "extra": "35838 times\n4 procs"
          },
          {
            "name": "BenchmarkReadHook_PreToolIntercept (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 45,
            "unit": "allocs/op",
            "extra": "35838 times\n4 procs"
          },
          {
            "name": "BenchmarkGrep_Grouped (github.com/alex60217101990/terse/internal/hook)",
            "value": 581708,
            "unit": "ns/op\t  170639 B/op\t     155 allocs/op",
            "extra": "2006 times\n4 procs"
          },
          {
            "name": "BenchmarkGrep_Grouped (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 581708,
            "unit": "ns/op",
            "extra": "2006 times\n4 procs"
          },
          {
            "name": "BenchmarkGrep_Grouped (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 170639,
            "unit": "B/op",
            "extra": "2006 times\n4 procs"
          },
          {
            "name": "BenchmarkGrep_Grouped (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 155,
            "unit": "allocs/op",
            "extra": "2006 times\n4 procs"
          },
          {
            "name": "BenchmarkGrep_Grouped (github.com/alex60217101990/terse/internal/hook)",
            "value": 580251,
            "unit": "ns/op\t  170661 B/op\t     155 allocs/op",
            "extra": "2074 times\n4 procs"
          },
          {
            "name": "BenchmarkGrep_Grouped (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 580251,
            "unit": "ns/op",
            "extra": "2074 times\n4 procs"
          },
          {
            "name": "BenchmarkGrep_Grouped (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 170661,
            "unit": "B/op",
            "extra": "2074 times\n4 procs"
          },
          {
            "name": "BenchmarkGrep_Grouped (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 155,
            "unit": "allocs/op",
            "extra": "2074 times\n4 procs"
          },
          {
            "name": "BenchmarkGrep_Grouped (github.com/alex60217101990/terse/internal/hook)",
            "value": 584527,
            "unit": "ns/op\t  170693 B/op\t     155 allocs/op",
            "extra": "1719 times\n4 procs"
          },
          {
            "name": "BenchmarkGrep_Grouped (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 584527,
            "unit": "ns/op",
            "extra": "1719 times\n4 procs"
          },
          {
            "name": "BenchmarkGrep_Grouped (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 170693,
            "unit": "B/op",
            "extra": "1719 times\n4 procs"
          },
          {
            "name": "BenchmarkGrep_Grouped (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 155,
            "unit": "allocs/op",
            "extra": "1719 times\n4 procs"
          },
          {
            "name": "BenchmarkGrep_Grouped (github.com/alex60217101990/terse/internal/hook)",
            "value": 573256,
            "unit": "ns/op\t  170642 B/op\t     155 allocs/op",
            "extra": "2052 times\n4 procs"
          },
          {
            "name": "BenchmarkGrep_Grouped (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 573256,
            "unit": "ns/op",
            "extra": "2052 times\n4 procs"
          },
          {
            "name": "BenchmarkGrep_Grouped (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 170642,
            "unit": "B/op",
            "extra": "2052 times\n4 procs"
          },
          {
            "name": "BenchmarkGrep_Grouped (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 155,
            "unit": "allocs/op",
            "extra": "2052 times\n4 procs"
          },
          {
            "name": "BenchmarkGrep_Grouped (github.com/alex60217101990/terse/internal/hook)",
            "value": 577339,
            "unit": "ns/op\t  170658 B/op\t     155 allocs/op",
            "extra": "2050 times\n4 procs"
          },
          {
            "name": "BenchmarkGrep_Grouped (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 577339,
            "unit": "ns/op",
            "extra": "2050 times\n4 procs"
          },
          {
            "name": "BenchmarkGrep_Grouped (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 170658,
            "unit": "B/op",
            "extra": "2050 times\n4 procs"
          },
          {
            "name": "BenchmarkGrep_Grouped (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 155,
            "unit": "allocs/op",
            "extra": "2050 times\n4 procs"
          },
          {
            "name": "BenchmarkGrep_Grouped (github.com/alex60217101990/terse/internal/hook)",
            "value": 579034,
            "unit": "ns/op\t  170627 B/op\t     155 allocs/op",
            "extra": "2052 times\n4 procs"
          },
          {
            "name": "BenchmarkGrep_Grouped (github.com/alex60217101990/terse/internal/hook) - ns/op",
            "value": 579034,
            "unit": "ns/op",
            "extra": "2052 times\n4 procs"
          },
          {
            "name": "BenchmarkGrep_Grouped (github.com/alex60217101990/terse/internal/hook) - B/op",
            "value": 170627,
            "unit": "B/op",
            "extra": "2052 times\n4 procs"
          },
          {
            "name": "BenchmarkGrep_Grouped (github.com/alex60217101990/terse/internal/hook) - allocs/op",
            "value": 155,
            "unit": "allocs/op",
            "extra": "2052 times\n4 procs"
          },
          {
            "name": "BenchmarkFlushDirty (github.com/alex60217101990/terse/internal/hookcore)",
            "value": 4821116,
            "unit": "ns/op\t   14819 B/op\t     227 allocs/op",
            "extra": "507 times\n4 procs"
          },
          {
            "name": "BenchmarkFlushDirty (github.com/alex60217101990/terse/internal/hookcore) - ns/op",
            "value": 4821116,
            "unit": "ns/op",
            "extra": "507 times\n4 procs"
          },
          {
            "name": "BenchmarkFlushDirty (github.com/alex60217101990/terse/internal/hookcore) - B/op",
            "value": 14819,
            "unit": "B/op",
            "extra": "507 times\n4 procs"
          },
          {
            "name": "BenchmarkFlushDirty (github.com/alex60217101990/terse/internal/hookcore) - allocs/op",
            "value": 227,
            "unit": "allocs/op",
            "extra": "507 times\n4 procs"
          },
          {
            "name": "BenchmarkFlushDirty (github.com/alex60217101990/terse/internal/hookcore)",
            "value": 5457444,
            "unit": "ns/op\t   15898 B/op\t     228 allocs/op",
            "extra": "220 times\n4 procs"
          },
          {
            "name": "BenchmarkFlushDirty (github.com/alex60217101990/terse/internal/hookcore) - ns/op",
            "value": 5457444,
            "unit": "ns/op",
            "extra": "220 times\n4 procs"
          },
          {
            "name": "BenchmarkFlushDirty (github.com/alex60217101990/terse/internal/hookcore) - B/op",
            "value": 15898,
            "unit": "B/op",
            "extra": "220 times\n4 procs"
          },
          {
            "name": "BenchmarkFlushDirty (github.com/alex60217101990/terse/internal/hookcore) - allocs/op",
            "value": 228,
            "unit": "allocs/op",
            "extra": "220 times\n4 procs"
          },
          {
            "name": "BenchmarkFlushDirty (github.com/alex60217101990/terse/internal/hookcore)",
            "value": 5442387,
            "unit": "ns/op\t   15735 B/op\t     228 allocs/op",
            "extra": "219 times\n4 procs"
          },
          {
            "name": "BenchmarkFlushDirty (github.com/alex60217101990/terse/internal/hookcore) - ns/op",
            "value": 5442387,
            "unit": "ns/op",
            "extra": "219 times\n4 procs"
          },
          {
            "name": "BenchmarkFlushDirty (github.com/alex60217101990/terse/internal/hookcore) - B/op",
            "value": 15735,
            "unit": "B/op",
            "extra": "219 times\n4 procs"
          },
          {
            "name": "BenchmarkFlushDirty (github.com/alex60217101990/terse/internal/hookcore) - allocs/op",
            "value": 228,
            "unit": "allocs/op",
            "extra": "219 times\n4 procs"
          },
          {
            "name": "BenchmarkFlushDirty (github.com/alex60217101990/terse/internal/hookcore)",
            "value": 5364646,
            "unit": "ns/op\t   15769 B/op\t     228 allocs/op",
            "extra": "222 times\n4 procs"
          },
          {
            "name": "BenchmarkFlushDirty (github.com/alex60217101990/terse/internal/hookcore) - ns/op",
            "value": 5364646,
            "unit": "ns/op",
            "extra": "222 times\n4 procs"
          },
          {
            "name": "BenchmarkFlushDirty (github.com/alex60217101990/terse/internal/hookcore) - B/op",
            "value": 15769,
            "unit": "B/op",
            "extra": "222 times\n4 procs"
          },
          {
            "name": "BenchmarkFlushDirty (github.com/alex60217101990/terse/internal/hookcore) - allocs/op",
            "value": 228,
            "unit": "allocs/op",
            "extra": "222 times\n4 procs"
          },
          {
            "name": "BenchmarkFlushDirty (github.com/alex60217101990/terse/internal/hookcore)",
            "value": 5436400,
            "unit": "ns/op\t   14079 B/op\t     228 allocs/op",
            "extra": "219 times\n4 procs"
          },
          {
            "name": "BenchmarkFlushDirty (github.com/alex60217101990/terse/internal/hookcore) - ns/op",
            "value": 5436400,
            "unit": "ns/op",
            "extra": "219 times\n4 procs"
          },
          {
            "name": "BenchmarkFlushDirty (github.com/alex60217101990/terse/internal/hookcore) - B/op",
            "value": 14079,
            "unit": "B/op",
            "extra": "219 times\n4 procs"
          },
          {
            "name": "BenchmarkFlushDirty (github.com/alex60217101990/terse/internal/hookcore) - allocs/op",
            "value": 228,
            "unit": "allocs/op",
            "extra": "219 times\n4 procs"
          },
          {
            "name": "BenchmarkFlushDirty (github.com/alex60217101990/terse/internal/hookcore)",
            "value": 5413924,
            "unit": "ns/op\t   17548 B/op\t     228 allocs/op",
            "extra": "220 times\n4 procs"
          },
          {
            "name": "BenchmarkFlushDirty (github.com/alex60217101990/terse/internal/hookcore) - ns/op",
            "value": 5413924,
            "unit": "ns/op",
            "extra": "220 times\n4 procs"
          },
          {
            "name": "BenchmarkFlushDirty (github.com/alex60217101990/terse/internal/hookcore) - B/op",
            "value": 17548,
            "unit": "B/op",
            "extra": "220 times\n4 procs"
          },
          {
            "name": "BenchmarkFlushDirty (github.com/alex60217101990/terse/internal/hookcore) - allocs/op",
            "value": 228,
            "unit": "allocs/op",
            "extra": "220 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSession (github.com/alex60217101990/terse/internal/hookcore)",
            "value": 2497,
            "unit": "ns/op\t    8304 B/op\t       5 allocs/op",
            "extra": "479019 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSession (github.com/alex60217101990/terse/internal/hookcore) - ns/op",
            "value": 2497,
            "unit": "ns/op",
            "extra": "479019 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSession (github.com/alex60217101990/terse/internal/hookcore) - B/op",
            "value": 8304,
            "unit": "B/op",
            "extra": "479019 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSession (github.com/alex60217101990/terse/internal/hookcore) - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "479019 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSession (github.com/alex60217101990/terse/internal/hookcore)",
            "value": 2435,
            "unit": "ns/op\t    8304 B/op\t       5 allocs/op",
            "extra": "487756 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSession (github.com/alex60217101990/terse/internal/hookcore) - ns/op",
            "value": 2435,
            "unit": "ns/op",
            "extra": "487756 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSession (github.com/alex60217101990/terse/internal/hookcore) - B/op",
            "value": 8304,
            "unit": "B/op",
            "extra": "487756 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSession (github.com/alex60217101990/terse/internal/hookcore) - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "487756 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSession (github.com/alex60217101990/terse/internal/hookcore)",
            "value": 2445,
            "unit": "ns/op\t    8304 B/op\t       5 allocs/op",
            "extra": "491943 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSession (github.com/alex60217101990/terse/internal/hookcore) - ns/op",
            "value": 2445,
            "unit": "ns/op",
            "extra": "491943 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSession (github.com/alex60217101990/terse/internal/hookcore) - B/op",
            "value": 8304,
            "unit": "B/op",
            "extra": "491943 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSession (github.com/alex60217101990/terse/internal/hookcore) - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "491943 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSession (github.com/alex60217101990/terse/internal/hookcore)",
            "value": 2443,
            "unit": "ns/op\t    8304 B/op\t       5 allocs/op",
            "extra": "492555 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSession (github.com/alex60217101990/terse/internal/hookcore) - ns/op",
            "value": 2443,
            "unit": "ns/op",
            "extra": "492555 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSession (github.com/alex60217101990/terse/internal/hookcore) - B/op",
            "value": 8304,
            "unit": "B/op",
            "extra": "492555 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSession (github.com/alex60217101990/terse/internal/hookcore) - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "492555 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSession (github.com/alex60217101990/terse/internal/hookcore)",
            "value": 2440,
            "unit": "ns/op\t    8304 B/op\t       5 allocs/op",
            "extra": "495802 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSession (github.com/alex60217101990/terse/internal/hookcore) - ns/op",
            "value": 2440,
            "unit": "ns/op",
            "extra": "495802 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSession (github.com/alex60217101990/terse/internal/hookcore) - B/op",
            "value": 8304,
            "unit": "B/op",
            "extra": "495802 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSession (github.com/alex60217101990/terse/internal/hookcore) - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "495802 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSession (github.com/alex60217101990/terse/internal/hookcore)",
            "value": 2455,
            "unit": "ns/op\t    8304 B/op\t       5 allocs/op",
            "extra": "461132 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSession (github.com/alex60217101990/terse/internal/hookcore) - ns/op",
            "value": 2455,
            "unit": "ns/op",
            "extra": "461132 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSession (github.com/alex60217101990/terse/internal/hookcore) - B/op",
            "value": 8304,
            "unit": "B/op",
            "extra": "461132 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSession (github.com/alex60217101990/terse/internal/hookcore) - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "461132 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveSession (github.com/alex60217101990/terse/internal/hookcore)",
            "value": 63.24,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "18922167 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveSession (github.com/alex60217101990/terse/internal/hookcore) - ns/op",
            "value": 63.24,
            "unit": "ns/op",
            "extra": "18922167 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveSession (github.com/alex60217101990/terse/internal/hookcore) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "18922167 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveSession (github.com/alex60217101990/terse/internal/hookcore) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "18922167 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveSession (github.com/alex60217101990/terse/internal/hookcore)",
            "value": 64.72,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "18928638 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveSession (github.com/alex60217101990/terse/internal/hookcore) - ns/op",
            "value": 64.72,
            "unit": "ns/op",
            "extra": "18928638 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveSession (github.com/alex60217101990/terse/internal/hookcore) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "18928638 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveSession (github.com/alex60217101990/terse/internal/hookcore) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "18928638 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveSession (github.com/alex60217101990/terse/internal/hookcore)",
            "value": 63.98,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "15955628 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveSession (github.com/alex60217101990/terse/internal/hookcore) - ns/op",
            "value": 63.98,
            "unit": "ns/op",
            "extra": "15955628 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveSession (github.com/alex60217101990/terse/internal/hookcore) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "15955628 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveSession (github.com/alex60217101990/terse/internal/hookcore) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "15955628 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveSession (github.com/alex60217101990/terse/internal/hookcore)",
            "value": 63.17,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "18931478 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveSession (github.com/alex60217101990/terse/internal/hookcore) - ns/op",
            "value": 63.17,
            "unit": "ns/op",
            "extra": "18931478 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveSession (github.com/alex60217101990/terse/internal/hookcore) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "18931478 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveSession (github.com/alex60217101990/terse/internal/hookcore) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "18931478 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveSession (github.com/alex60217101990/terse/internal/hookcore)",
            "value": 63.17,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "18936877 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveSession (github.com/alex60217101990/terse/internal/hookcore) - ns/op",
            "value": 63.17,
            "unit": "ns/op",
            "extra": "18936877 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveSession (github.com/alex60217101990/terse/internal/hookcore) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "18936877 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveSession (github.com/alex60217101990/terse/internal/hookcore) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "18936877 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveSession (github.com/alex60217101990/terse/internal/hookcore)",
            "value": 63.21,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "18968678 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveSession (github.com/alex60217101990/terse/internal/hookcore) - ns/op",
            "value": 63.21,
            "unit": "ns/op",
            "extra": "18968678 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveSession (github.com/alex60217101990/terse/internal/hookcore) - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "18968678 times\n4 procs"
          },
          {
            "name": "BenchmarkSaveSession (github.com/alex60217101990/terse/internal/hookcore) - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "18968678 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSessionReloadMiss (github.com/alex60217101990/terse/internal/hookcore)",
            "value": 78717,
            "unit": "ns/op\t  108113 B/op\t      24 allocs/op",
            "extra": "15198 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSessionReloadMiss (github.com/alex60217101990/terse/internal/hookcore) - ns/op",
            "value": 78717,
            "unit": "ns/op",
            "extra": "15198 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSessionReloadMiss (github.com/alex60217101990/terse/internal/hookcore) - B/op",
            "value": 108113,
            "unit": "B/op",
            "extra": "15198 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSessionReloadMiss (github.com/alex60217101990/terse/internal/hookcore) - allocs/op",
            "value": 24,
            "unit": "allocs/op",
            "extra": "15198 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSessionReloadMiss (github.com/alex60217101990/terse/internal/hookcore)",
            "value": 78120,
            "unit": "ns/op\t  108114 B/op\t      24 allocs/op",
            "extra": "15344 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSessionReloadMiss (github.com/alex60217101990/terse/internal/hookcore) - ns/op",
            "value": 78120,
            "unit": "ns/op",
            "extra": "15344 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSessionReloadMiss (github.com/alex60217101990/terse/internal/hookcore) - B/op",
            "value": 108114,
            "unit": "B/op",
            "extra": "15344 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSessionReloadMiss (github.com/alex60217101990/terse/internal/hookcore) - allocs/op",
            "value": 24,
            "unit": "allocs/op",
            "extra": "15344 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSessionReloadMiss (github.com/alex60217101990/terse/internal/hookcore)",
            "value": 72459,
            "unit": "ns/op\t  108114 B/op\t      24 allocs/op",
            "extra": "16748 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSessionReloadMiss (github.com/alex60217101990/terse/internal/hookcore) - ns/op",
            "value": 72459,
            "unit": "ns/op",
            "extra": "16748 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSessionReloadMiss (github.com/alex60217101990/terse/internal/hookcore) - B/op",
            "value": 108114,
            "unit": "B/op",
            "extra": "16748 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSessionReloadMiss (github.com/alex60217101990/terse/internal/hookcore) - allocs/op",
            "value": 24,
            "unit": "allocs/op",
            "extra": "16748 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSessionReloadMiss (github.com/alex60217101990/terse/internal/hookcore)",
            "value": 78079,
            "unit": "ns/op\t  108113 B/op\t      24 allocs/op",
            "extra": "15404 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSessionReloadMiss (github.com/alex60217101990/terse/internal/hookcore) - ns/op",
            "value": 78079,
            "unit": "ns/op",
            "extra": "15404 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSessionReloadMiss (github.com/alex60217101990/terse/internal/hookcore) - B/op",
            "value": 108113,
            "unit": "B/op",
            "extra": "15404 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSessionReloadMiss (github.com/alex60217101990/terse/internal/hookcore) - allocs/op",
            "value": 24,
            "unit": "allocs/op",
            "extra": "15404 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSessionReloadMiss (github.com/alex60217101990/terse/internal/hookcore)",
            "value": 77390,
            "unit": "ns/op\t  108114 B/op\t      24 allocs/op",
            "extra": "15606 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSessionReloadMiss (github.com/alex60217101990/terse/internal/hookcore) - ns/op",
            "value": 77390,
            "unit": "ns/op",
            "extra": "15606 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSessionReloadMiss (github.com/alex60217101990/terse/internal/hookcore) - B/op",
            "value": 108114,
            "unit": "B/op",
            "extra": "15606 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSessionReloadMiss (github.com/alex60217101990/terse/internal/hookcore) - allocs/op",
            "value": 24,
            "unit": "allocs/op",
            "extra": "15606 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSessionReloadMiss (github.com/alex60217101990/terse/internal/hookcore)",
            "value": 77961,
            "unit": "ns/op\t  108114 B/op\t      24 allocs/op",
            "extra": "15435 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSessionReloadMiss (github.com/alex60217101990/terse/internal/hookcore) - ns/op",
            "value": 77961,
            "unit": "ns/op",
            "extra": "15435 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSessionReloadMiss (github.com/alex60217101990/terse/internal/hookcore) - B/op",
            "value": 108114,
            "unit": "B/op",
            "extra": "15435 times\n4 procs"
          },
          {
            "name": "BenchmarkLoadSessionReloadMiss (github.com/alex60217101990/terse/internal/hookcore) - allocs/op",
            "value": 24,
            "unit": "allocs/op",
            "extra": "15435 times\n4 procs"
          },
          {
            "name": "BenchmarkToolResponseUnmarshalContent10KB (github.com/alex60217101990/terse/internal/protocol)",
            "value": 115761,
            "unit": "ns/op\t   23609 B/op\t       8 allocs/op",
            "extra": "10000 times\n4 procs"
          },
          {
            "name": "BenchmarkToolResponseUnmarshalContent10KB (github.com/alex60217101990/terse/internal/protocol) - ns/op",
            "value": 115761,
            "unit": "ns/op",
            "extra": "10000 times\n4 procs"
          },
          {
            "name": "BenchmarkToolResponseUnmarshalContent10KB (github.com/alex60217101990/terse/internal/protocol) - B/op",
            "value": 23609,
            "unit": "B/op",
            "extra": "10000 times\n4 procs"
          },
          {
            "name": "BenchmarkToolResponseUnmarshalContent10KB (github.com/alex60217101990/terse/internal/protocol) - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "10000 times\n4 procs"
          },
          {
            "name": "BenchmarkToolResponseUnmarshalContent10KB (github.com/alex60217101990/terse/internal/protocol)",
            "value": 116655,
            "unit": "ns/op\t   23608 B/op\t       8 allocs/op",
            "extra": "10000 times\n4 procs"
          },
          {
            "name": "BenchmarkToolResponseUnmarshalContent10KB (github.com/alex60217101990/terse/internal/protocol) - ns/op",
            "value": 116655,
            "unit": "ns/op",
            "extra": "10000 times\n4 procs"
          },
          {
            "name": "BenchmarkToolResponseUnmarshalContent10KB (github.com/alex60217101990/terse/internal/protocol) - B/op",
            "value": 23608,
            "unit": "B/op",
            "extra": "10000 times\n4 procs"
          },
          {
            "name": "BenchmarkToolResponseUnmarshalContent10KB (github.com/alex60217101990/terse/internal/protocol) - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "10000 times\n4 procs"
          },
          {
            "name": "BenchmarkToolResponseUnmarshalContent10KB (github.com/alex60217101990/terse/internal/protocol)",
            "value": 115602,
            "unit": "ns/op\t   23608 B/op\t       8 allocs/op",
            "extra": "10000 times\n4 procs"
          },
          {
            "name": "BenchmarkToolResponseUnmarshalContent10KB (github.com/alex60217101990/terse/internal/protocol) - ns/op",
            "value": 115602,
            "unit": "ns/op",
            "extra": "10000 times\n4 procs"
          },
          {
            "name": "BenchmarkToolResponseUnmarshalContent10KB (github.com/alex60217101990/terse/internal/protocol) - B/op",
            "value": 23608,
            "unit": "B/op",
            "extra": "10000 times\n4 procs"
          },
          {
            "name": "BenchmarkToolResponseUnmarshalContent10KB (github.com/alex60217101990/terse/internal/protocol) - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "10000 times\n4 procs"
          },
          {
            "name": "BenchmarkToolResponseUnmarshalContent10KB (github.com/alex60217101990/terse/internal/protocol)",
            "value": 117312,
            "unit": "ns/op\t   23608 B/op\t       8 allocs/op",
            "extra": "10000 times\n4 procs"
          },
          {
            "name": "BenchmarkToolResponseUnmarshalContent10KB (github.com/alex60217101990/terse/internal/protocol) - ns/op",
            "value": 117312,
            "unit": "ns/op",
            "extra": "10000 times\n4 procs"
          },
          {
            "name": "BenchmarkToolResponseUnmarshalContent10KB (github.com/alex60217101990/terse/internal/protocol) - B/op",
            "value": 23608,
            "unit": "B/op",
            "extra": "10000 times\n4 procs"
          },
          {
            "name": "BenchmarkToolResponseUnmarshalContent10KB (github.com/alex60217101990/terse/internal/protocol) - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "10000 times\n4 procs"
          },
          {
            "name": "BenchmarkToolResponseUnmarshalContent10KB (github.com/alex60217101990/terse/internal/protocol)",
            "value": 121893,
            "unit": "ns/op\t   23608 B/op\t       8 allocs/op",
            "extra": "10000 times\n4 procs"
          },
          {
            "name": "BenchmarkToolResponseUnmarshalContent10KB (github.com/alex60217101990/terse/internal/protocol) - ns/op",
            "value": 121893,
            "unit": "ns/op",
            "extra": "10000 times\n4 procs"
          },
          {
            "name": "BenchmarkToolResponseUnmarshalContent10KB (github.com/alex60217101990/terse/internal/protocol) - B/op",
            "value": 23608,
            "unit": "B/op",
            "extra": "10000 times\n4 procs"
          },
          {
            "name": "BenchmarkToolResponseUnmarshalContent10KB (github.com/alex60217101990/terse/internal/protocol) - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "10000 times\n4 procs"
          },
          {
            "name": "BenchmarkToolResponseUnmarshalContent10KB (github.com/alex60217101990/terse/internal/protocol)",
            "value": 115812,
            "unit": "ns/op\t   23608 B/op\t       8 allocs/op",
            "extra": "8733 times\n4 procs"
          },
          {
            "name": "BenchmarkToolResponseUnmarshalContent10KB (github.com/alex60217101990/terse/internal/protocol) - ns/op",
            "value": 115812,
            "unit": "ns/op",
            "extra": "8733 times\n4 procs"
          },
          {
            "name": "BenchmarkToolResponseUnmarshalContent10KB (github.com/alex60217101990/terse/internal/protocol) - B/op",
            "value": 23608,
            "unit": "B/op",
            "extra": "8733 times\n4 procs"
          },
          {
            "name": "BenchmarkToolResponseUnmarshalContent10KB (github.com/alex60217101990/terse/internal/protocol) - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "8733 times\n4 procs"
          }
        ]
      }
    ]
  }
}