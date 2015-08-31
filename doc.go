// Katana package provides a dependency injection solution driven by constructor functions
//
// A constructor function is essentially a function that knows how to create an instance
// of a particular type, taking arguments representing that type's dependencies -- if any.
//
// Katana will call the appropriate constructor function when a new instance of a type is
// requested resolving any dependency that constructor may have by recursively calling
// their constructor functions. Cyclic dependencies are detected by keeping a stack of the
// dependency tree resolution
package katana
