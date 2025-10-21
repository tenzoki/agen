Summary by Location

  | File                      | Setting           | Line        | Default  |
  |---------------------------|-------------------|-------------|----------|
  | alfa.yaml                 | debug             | 1           | false    |
  | cellorg.yaml              | debug             | 5           | false    |
  | cellorg.yaml              | support.debug     | 16          | false    |
  | cellorg.yaml              | broker.debug      | 23          | false    |
  | plan-execute-verify.yaml  | debug             | 4           | true     |
  | plan-execute-verify.yaml  | log_agent_output  | 5           | false    |
  | plan-execute-verify.yaml  | log_intermediates | 6           | true     |
  | Most pipelines            | debug             | ~5          | true     |
  | Most services             | debug             | ~5          | true     |
  | ocr-production.yaml       | debug             | 5           | false ⚠️ |
  | fast-analysis.yaml        | debug             | 5           | false ⚠️ |
  | pool.yaml (anonymization) | enable_debug      | 528,548,566 | false    |

  Quick Toggle Points

  To reduce log noise globally:
  1. alfa.yaml:1 - Alfa orchestrator
  2. cellorg.yaml:5,16,23 - Cellorg services
  3. plan-execute-verify.yaml:4 - PEV cell
  4. plan-execute-verify.yaml:5 - Agent output logging

  Most effective for PEV: Set plan-execute-verify.yaml lines 4-6 to false.
