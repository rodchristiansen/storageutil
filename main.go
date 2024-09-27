package main

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
)

// Function to run "du -sh" for each user directory
func getDirSize(dir string) (string, error) {
    cmd := exec.Command("du", "-sh", dir)
    output, err := cmd.Output()
    if err != nil {
        return "", err
    }
    size := strings.Fields(string(output))[0]
    return size, nil
}

// Function to write to plist using the defaults command
func writeToPlist(username, size string) error {
    plistLoc := "/Library/Management/ca.ecuad.macadmins.UserStorage.plist"
    plistKey := fmt.Sprintf("%s", username)  // Simplified to just use the username
    cmd := exec.Command("sudo", "defaults", "write", plistLoc, plistKey, "-string", size)

    if err := cmd.Run(); err != nil {
        return err
    }

    // Ensure the plist file has the correct permissions (chmod 755)
    if err := os.Chmod(plistLoc, 0755); err != nil {
        return fmt.Errorf("failed to set permissions on plist: %v", err)
    }

    return nil
}

// Function to clear the plist file before adding new data
func clearPlist() error {
    plistLoc := "/Library/Management/ca.ecuad.macadmins.UserStorage.plist"
    
    // Check if the plist file exists
    if _, err := os.Stat(plistLoc); os.IsNotExist(err) {
        // If plist does not exist, no need to delete it
        fmt.Println("Plist does not exist, nothing to clear")
        return nil
    }

    // Run the defaults delete command
    cmd := exec.Command("sudo", "defaults", "delete", plistLoc)
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("failed to clear plist: %v", err)
    }

    return nil
}

func main() {
    userDir := "/Users"

    // Clear the plist before collecting new data
    if err := clearPlist(); err != nil {
        fmt.Printf("Error clearing plist: %v\n", err)
        return
    }

    // List directories directly inside /Users
    entries, err := os.ReadDir(userDir)
    if err != nil {
        fmt.Printf("Error reading /Users: %v\n", err)
        return
    }

    // Loop through the contents of /Users, only processing directories
    for _, entry := range entries {
        if entry.IsDir() {
            username := entry.Name()
            // Ignore root, Shared, and other system directories
            if username == "root" || username == "Shared" || username == "Library" {
                continue
            }

            // Construct the full path to the user's home directory
            homeDir := filepath.Join(userDir, username)

            // Calculate the total size of the user's home directory
            size, err := getDirSize(homeDir)
            if err != nil {
                fmt.Printf("Error getting size for %s: %v\n", username, err)
                continue
            }

            fmt.Printf("Username: %s, Home Folder Size: %s\n", username, size)

            // Write the user's home directory size to the plist
            if err := writeToPlist(username, size); err != nil {
                fmt.Printf("Error writing plist for %s: %v\n", username, err)
            }
        }
    }
}
