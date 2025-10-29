package handlers

import (
    "context"
    "fmt"
    "log"

	"github.com/balaji/hello/ent"
    "github.com/balaji/hello/pkg/application"
)

func Persist(ctx context.Context, client *ent.Client, ad *application.ApplicationDescription) error {
    //ad := ads[0]
    log.Println("[CO] persisting app desc 1:", ad)
    //return nil

    // save catlog info 
    _, err := client.ApplicationDesc.
        Create().
        SetID(ad.Metadata.ID).
        SetName(ad.Metadata.Name).
        SetVersion(ad.Metadata.Version).
        SetTags(ad.Metadata.Catalog.Application.Tags).
        SetTagLine(ad.Metadata.Catalog.Application.Tagline).
        SetVendor(ad.Metadata.Catalog.Application.Site).
        SetCategory(ad.Metadata.Catalog.Application.Site).
        SetDescription(ad.Metadata.Catalog.Application.DescriptionFile).    
        SetIcon(ad.Metadata.Catalog.Application.Icon).    
        SetSite(ad.Metadata.Catalog.Application.Site).    
        Save(ctx)        
    if err != nil { return err }

    // // Store 
    log.Println("number of dep profiles: ", len(ad.DeploymentProfiles))
    
    for _, spec := range ad.DeploymentProfiles {

        dpCreate := client.DeploymentProfile.
            Create().
            SetID(spec.ID).
            SetType(spec.Type).
            SetAppID(ad.Metadata.ID)
        
        if spec.Description != "" {
            dpCreate.SetDescription(spec.Description)
        }
        log.Println("ad.RequiredResources", spec.RequiredResources)
        if spec.RequiredResources != nil {
            log.Println("CPU Cores:", spec.RequiredResources.CPU.Cores)
            // for _, p := range ad.RequiredResources.Peripherals {
            //     log.Println("Peripheral:", p)
            // }
            dpCreate.
                SetCPUCores(spec.RequiredResources.CPU.Cores).
                SetMemory(spec.RequiredResources.Memory).
                SetStorage(spec.RequiredResources.Storage).
                SetCPUArchitectures(spec.RequiredResources.CPU.Architectures).
                SetPeripherals(peripheralsToMap(spec.RequiredResources.Peripherals)).
                SetInterfaces(interfacesToMap(spec.RequiredResources.Interfaces))               
        } else {
            log.Println("Skipping RequiredResources, it is nil")
            //continue
        }
        _, err := dpCreate.Save(ctx)
        if err != nil {
            return fmt.Errorf("failed to create deployment profile: %w", err)
        }
        if err != nil { return err }

        // components
        for _, component := range spec.Components {
            _, err := client.Component.
                Create().
                SetName(component.Name).
                SetDeploymentProfileID(spec.ID).
                SetProperties(application.ComponentProperties{
                    Repository:     component.Properties.Repository,
                    Revision:       component.Properties.Revision,
                    Wait:           component.Properties.Wait,
                    Timeout:        component.Properties.Timeout,
                    PackageLocation: component.Properties.PackageLocation,
                    KeyLocation:     component.Properties.KeyLocation,  
                }).  
                Save(ctx)
            if err != nil {
                return fmt.Errorf("failed to create component: %w", err)
            }
        }
    }  

    return nil
}

func peripheralsToMap(peripherals []application.Peripheral) []map[string]interface{} {
    result := make([]map[string]interface{}, len(peripherals))
    for i, p := range peripherals {
        result[i] = map[string]interface{}{
            "manufacturer": p.Manufacturer,
            "type": p.Type,
            "model": p.Model,
            // add more fields here if needed
        }
    }
    return result
}

func interfacesToMap(interfaces []application.Interface) []map[string]interface{} {
    result := make([]map[string]interface{}, len(interfaces))
    for i, iface := range interfaces {
        result[i] = map[string]interface{}{
            "type": iface.Type,
            //"protocol": iface.Protocol,
            // add more fields here if needed
        }
    }
    return result
}
