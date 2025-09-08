graph TD
    N1["flux-system<br/>📁 flux-kustomization"]
    N2["backend<br/>🚀 helm-release"]
    N3["frontend<br/>🚀 helm-release"]
    N4["postgres<br/>🚀 helm-release"]

    %% Orphaned Resources
    N5["production<br/>📄 kubernetes-resource"]
    N6["postgres<br/>📄 kubernetes-resource"]
    N7["backend<br/>📄 kubernetes-resource"]
    N8["frontend<br/>📄 kubernetes-resource"]
    N9["orphaned-config<br/>📄 kubernetes-resource"]

    %% Styling
    classDef valid fill:#90EE90,stroke:#333,stroke-width:2px
    classDef orphaned fill:#FFB6C1,stroke:#333,stroke-width:2px
    classDef error fill:#FF6B6B,stroke:#333,stroke-width:2px
    classDef warning fill:#FFE4B5,stroke:#333,stroke-width:2px

    %% Apply styles
    class N9 valid
    class N1 valid
    class N5 valid
    class N6 valid
    class N8 valid
    class N2 valid
    class N3 valid
    class N4 valid
    class N7 valid
    class N5 orphaned
    class N6 orphaned
    class N7 orphaned
    class N8 orphaned
    class N9 orphaned