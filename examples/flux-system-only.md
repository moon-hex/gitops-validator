graph TD
    N1["flux-system<br/>📁 flux-kustomization"]

    %% Orphaned Resources
    N2["production<br/>📄 kubernetes-resource"]
    N3["postgres<br/>🚀 helm-release"]
    N4["postgres<br/>📄 kubernetes-resource"]
    N5["orphaned-config<br/>📄 kubernetes-resource"]
    N6["backend<br/>🚀 helm-release"]
    N7["frontend<br/>🚀 helm-release"]
    N8["frontend<br/>📄 kubernetes-resource"]
    N9["backend<br/>📄 kubernetes-resource"]

    %% Styling
    classDef valid fill:#2E8B57,stroke:#1F5F3F,stroke-width:3px,color:#FFFFFF
    classDef orphaned fill:#DC143C,stroke:#8B0000,stroke-width:3px,color:#FFFFFF
    classDef error fill:#B22222,stroke:#8B0000,stroke-width:3px,color:#FFFFFF
    classDef warning fill:#FF8C00,stroke:#CC7000,stroke-width:3px,color:#FFFFFF

    %% Apply styles
    class N4 valid
    class N8 valid
    class N1 valid
    class N2 valid
    class N3 valid
    class N9 valid
    class N5 valid
    class N6 valid
    class N7 valid
    class N2 orphaned
    class N3 orphaned
    class N4 orphaned
    class N5 orphaned
    class N6 orphaned
    class N7 orphaned
    class N8 orphaned
    class N9 orphaned